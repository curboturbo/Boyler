package image

import (
	"boyler/internal/daemon/core"
	domain "boyler/internal/daemon/core"
	string_pkg "boyler/pkg/string"
	"boyler/pkg/files"
	"boyler/pkg/logger"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

)

type imageManager struct {
    imageDir string
    OperationSystem string
    Architecture string
}

func NewImageManager(imageDir string) ImageManager {
    return &imageManager{imageDir: imageDir, 
        OperationSystem: "linux", 
        Architecture: "amd64",
    }
}

func (i *imageManager) Extract(ctx context.Context, name string, unpackDir string) error {
	log := logger.FromContext(ctx)
	safeName := string_pkg.SanitizeImageName(name)
	unpackDir = filepath.Join(unpackDir,safeName, "rootfs")

	if err := os.MkdirAll(unpackDir, 0755); err != nil {
    	return fmt.Errorf("create rootfs: %w", err)
	}
	log.Debug("start extract image", "name", safeName)
	imageDir := filepath.Join(i.imageDir, safeName)
	num, err := readLayersInfo(imageDir)
	if err != nil {
		return fmt.Errorf("read layers info: %w", err)
	}
	for idx := 0; idx < num; idx++ {
		layerPath := filepath.Join(imageDir, fmt.Sprintf("layer_%d.tar.gz", idx))
		log.Debug("extracting layer", "path", layerPath)
		if err := files.Unzip(layerPath, unpackDir); err != nil {
			return fmt.Errorf("unzip layer %s: %w", layerPath, err)
		}
	}
	log.Info("image extracted", "layers", num)
	return nil
}

func (i *imageManager) IsExtracted(ctx context.Context, name string) bool {
    log := logger.FromContext(ctx)
	safeName := string_pkg.SanitizeImageName(name)
    log.Debug("check is image extractred", "name",safeName)
    rootfsPath := filepath.Join(i.imageDir, safeName, "rootfs")
    info, err := os.Stat(rootfsPath)
    if err != nil {
        if os.IsNotExist(err) {
            log.Warn("image not extracted yet","image",safeName)
        }else{
            log.Warn("failed to check if image is extracted","image",safeName)
        }
        return false
    }
    
    if !info.IsDir() {
        log.Warn("rootfs exists but is not a directory","image",safeName)
        return false
    }
    return true
}


func (i *imageManager) GetRootfsPath(name string) string {
    return filepath.Join(i.imageDir, name, "rootfs")
}

func (i *imageManager) Delete(ctx context.Context, name string) error {
    log := logger.FromContext(ctx)
    deletePath := filepath.Join(i.imageDir, name)
    _, err := os.Stat(deletePath)
    if err != nil {
        if os.IsNotExist(err) {
            log.Warn("image does not exist, nothing to delete","image", name)
            return nil
        }
        log.Warn("failed to check image existence","image",name)
        return err
    }
    err = os.RemoveAll(deletePath)
    if err != nil {
        log.Warn("failed to delete image", "image",name)
        return err
    }
    log.Info("image deleted successfully", "image", name)
    return nil
}

func (i *imageManager) Get(ctx context.Context, name string) (*domain.Image,error) {
    log := logger.FromContext(ctx)
	path := filepath.Join(i.imageDir, name, "meta.json")
	_, err := os.Stat(path)
	if err != nil{
		log.Warn("failed to find image","image",name,"path", path)
		return nil, err
	}
	data, err := os.ReadFile(path)
	if err != nil {
		log.Warn("failed to read metadata file", 
			"image", name,
			"path", path,
        )
		return nil, err
	}
	var metaData domain.Image
	err = json.Unmarshal(data, &metaData)
	if err != nil {
        return nil, err
	}
	return &metaData, nil
}


func (i *imageManager) List(ctx context.Context) ([]*domain.Image, error) {
	subDirs, err := os.ReadDir(i.imageDir)
	if err != nil{
        return []*domain.Image{}, err
	}
	var images []*domain.Image
	for _, dir := range subDirs{
		if dir.IsDir(){
			continue
		}
		name := dir.Name()
		image, err := i.Get(ctx, name)
		if err != nil{
			continue
		}else{
		images = append(images, image)
		}
	}
	return images, nil
}

func (i *imageManager) Pull(ctx context.Context, name string, ch chan *core.PullingEvent) error {
    defer close(ch)
    imagePuller := NewDockerHubPuller(Platform{
        OS: i.OperationSystem,
        Architecture: i.Architecture,
    }, ch)
    _, err := imagePuller.Pull(ctx, name, i.imageDir)
    if err != nil{
        return fmt.Errorf("Failed to fetch image: %v", err)
    }
	return nil
}


func readLayersInfo(dir string) (int, error) {
	path := filepath.Join(dir, layersInfoFileName)
	data, err := os.ReadFile(path)
	if err != nil {
		return 0, fmt.Errorf("read layers info: %w", err)
    }
	var info layersInfo
	if err := json.Unmarshal(data, &info); err != nil {
		return 0, fmt.Errorf("unmarshal layers info: %w", err)
	}
	return info.Num, nil
}