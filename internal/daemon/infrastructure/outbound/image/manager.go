package image

import (
	domain "boyler/internal/daemon/core"
	"boyler/pkg/files"
	"boyler/pkg/logger"
	"context"
	"encoding/json"
	"os"
	"path/filepath"
)

type imageManager struct {
    imageDir string
}

func NewImageManager(imageDir string) ImageManager {
    return &imageManager{imageDir: imageDir}
}

func (i *imageManager) Extract(ctx context.Context, name string, unpackDir string) error {
    log := logger.FromContext(ctx)
    log.Debug("start extract image", "name",name)
    archivePath := filepath.Join(i.imageDir, name, name+".tar.gz")
    err := files.Unzip(archivePath, unpackDir)
    if err != nil {
        return err
    }
    log.Info("image extracted")
    return nil
}

func (i *imageManager) IsExtracted(ctx context.Context, name string) bool {
    log := logger.FromContext(ctx)
    log.Debug("check is image extractred", "name",name)
    rootfsPath := filepath.Join(i.imageDir, name, "rootfs")
    info, err := os.Stat(rootfsPath)
    if err != nil {
        if os.IsNotExist(err) {
            log.Warn("image not extracted yet","image",name)
        }else{
            log.Warn("failed to check if image is extracted","image",name)
        }
        return false
    }
    
    if !info.IsDir() {
        log.Warn("rootfs exists but is not a directory","image",name)
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

func (i *imageManager)  Pull(ctx context.Context, name string) (*domain.Image, error) {
	// must download defined image from servers
	// save to boyler/images/{name}/{name.tar.gz}, meta.json
	// and return *domain.Image
	// perhaps download from dockerHub or something like it
	return &domain.Image{}, nil
}