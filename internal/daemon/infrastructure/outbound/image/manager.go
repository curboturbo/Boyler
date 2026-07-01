package image

import (
	domain "boyler/internal/daemon/core"
	"boyler/pkg/files"
	"encoding/json"
	"log/slog"
	"os"
	"path/filepath"
)

type imageManager struct {
    imageDir string
    logger   *slog.Logger
}

func NewImageManager(imageDir string, logger *slog.Logger) ImageManager {
    return &imageManager{imageDir: imageDir, logger:logger}
}

func (i *imageManager) Extract(name string, unpackDir string) error {
    archivePath := filepath.Join(i.imageDir, name, name+".tar.gz")
    i.logger.Info("extracting image",
        slog.String("image", name),
        slog.String("archive", archivePath),
    )
    err := files.Unzip(archivePath, unpackDir)
    if err != nil {
        i.logger.Error("failed to unzip archive",
            slog.String("image", name),
            slog.String("archive", archivePath),
            slog.Any("error", err),
        )
        return err
    }
    i.logger.Info("archive unpacked successfully", 
        slog.String("image", name),
        slog.String("path", unpackDir),
    )
    return nil
}

func (i *imageManager) IsExtracted(name string) bool {
    rootfsPath := filepath.Join(i.imageDir, name, "rootfs")
    info, err := os.Stat(rootfsPath)
    if err != nil {
        if os.IsNotExist(err) {
            i.logger.Debug("image not extracted yet", 
                slog.String("image", name))
        }else{
            i.logger.Error("failed to check if image is extracted",
                slog.String("image", name),
                slog.Any("error", err),
            )
        }
        return false
    }
    
    if !info.IsDir() {
        i.logger.Warn("rootfs exists but is not a directory",
            slog.String("image", name),
        )
        return false
    }
    
    i.logger.Debug("image is extracted", 
        slog.String("image", name))
    return true
}

func (i *imageManager) GetRootfsPath(name string) string {
    return filepath.Join(i.imageDir, name, "rootfs")
}

func (i *imageManager) Delete(name string) error {
    deletePath := filepath.Join(i.imageDir, name)
    _, err := os.Stat(deletePath)
    if err != nil {
        if os.IsNotExist(err) {
            i.logger.Warn("image does not exist, nothing to delete", 
                slog.String("image", name))
            return nil
        }
        i.logger.Error("failed to check image existence",
            slog.String("image", name),
            slog.Any("error", err),
        )
        return err
    }
    i.logger.Info("deleting image", slog.String("image", name))
    err = os.RemoveAll(deletePath)
    if err != nil {
        i.logger.Error("failed to delete image",
            slog.String("image", name),
            slog.Any("error", err),
        )
        return err
    }
    i.logger.Info("image deleted successfully", 
        slog.String("image", name))
    return nil
}

func (i *imageManager) Get(name string) (*domain.Image,error) {
	path := filepath.Join(i.imageDir, name, "meta.json")
	_, err := os.Stat(path)
	if err != nil{
		i.logger.Error("failed to find image",
		slog.String("image",name),
		slog.String("path", path),
	)
		return &domain.Image{}, err
	}

	data, err := os.ReadFile(path)
	if err != nil {
		i.logger.Error("failed to read metadata file", 
			slog.String("image", name),
			slog.String("path", path),
			slog.Any("error", err))
		return &domain.Image{}, err
	}
	var metaData domain.Image
	err = json.Unmarshal(data, &metaData)
	if err != nil {
		i.logger.Error("failed to parse metadata",
		slog.String("image", name),
		slog.Any("error", err),
		)
	return nil, err
	}
	i.logger.Debug("image loaded", 
		slog.String("image", name),
		slog.String("id", metaData.ID))
	return &metaData, nil
}


func (i *imageManager) List() ([]*domain.Image, error) {
	subDirs, err := os.ReadDir(i.imageDir)
	if err != nil{
		i.logger.Error("failed to read images directory",
		slog.String("path", i.imageDir),
	)
	}
	var images []*domain.Image
	for _, dir := range subDirs{
		if dir.IsDir(){
			continue
		}
		name := dir.Name()
		image, err := i.Get(name)
		if err != nil{
			continue
		}else{
		images = append(images, image)
		}
	}
	i.logger.Debug("images iterated", slog.Int("count", len(images)))
	return images, nil
}

func (i *imageManager)  Pull(name string) (*domain.Image, error) {
	// must download defined image from servers
	// save to boyler/images/{name}/{name.tar.gz}, meta.json
	// and return *domain.Image
	// perhaps download from dockerHub or something like it
	return &domain.Image{}, nil
}