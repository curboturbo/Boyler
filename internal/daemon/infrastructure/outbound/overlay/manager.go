package overlay

import (
	"boyler/pkg/logger"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"syscall"
	string_pkg "boyler/pkg/string"
)

type overlayManager struct {
	imageDir string // images directory {boyler/lib/images}
	containerDir string // containers directory {boyler/lib/containers}
}

func NewOverlayManager(imageDir string, containerDir string) VolumeManager{
	return &overlayManager{
		imageDir: imageDir,
		containerDir: containerDir,
	}
}

func (vm *overlayManager) CreateMountPoints(ctx context.Context, containerID string) error {
	log := logger.FromContext(ctx)
	log.Debug("start unpack create mount points","containerID",containerID)
	containerPath := filepath.Join(vm.containerDir, containerID)
	err := os.MkdirAll(containerPath, 0755)
	if err != nil {
		return err
	}
	merged := filepath.Join(containerPath, "merged")
	work := filepath.Join(containerPath, "work")
	upper := filepath.Join(containerPath,"upper")
	for _, path := range []string{merged, work, upper} {
		err = os.MkdirAll(path,0755)
		if err != nil{
			os.RemoveAll(containerPath)
			return err
		}
	}
	log.Info("mount directory created")
	return nil
}

func (vm *overlayManager) Mount(ctx context.Context, containerID string, imageName string) error {
	log := logger.FromContext(ctx)
	safeName := string_pkg.SanitizeImageName(imageName)
	log.Debug("start mount","containerID",containerID, "imageName", safeName)
	containerPath := filepath.Join(vm.containerDir, containerID)

	mergedDir := filepath.Join(containerPath, "merged")
	upperDir := filepath.Join(containerPath, "upper")
	workDir := filepath.Join(containerPath, "work")
	lowerDir := filepath.Join(vm.imageDir, safeName, "rootfs")

	if _, err := os.Stat(lowerDir); os.IsNotExist(err){
		return err
	}

	opts := fmt.Sprintf("lowerdir=%s,upperdir=%s,workdir=%s", filepath.Clean(lowerDir), filepath.Clean(upperDir), filepath.Clean(workDir))
	err := syscall.Mount("overlay", mergedDir, "overlay", 0, opts)

	if err != nil {
		return err
	}
	log.Info("mounting created")
	return nil
}


func (vm *overlayManager) Unmount(ctx context.Context,containerID string) error {
	mergedDir := filepath.Join(vm.containerDir, containerID, "merged")
	return syscall.Unmount(mergedDir, 0)
}


func (vm *overlayManager) Cleanup(ctx context.Context, containerID string) error {
	log := logger.FromContext(ctx)
	log.Debug("start cleanup", "containerID",containerID)
	delDir := filepath.Join(vm.containerDir, containerID)
	entries, err := os.ReadDir(delDir)
	if err != nil {return err}
	for _, entry := range entries {
		err := os.RemoveAll(filepath.Join(delDir, entry.Name()))
		if err != nil {return err}
	}
	if err := os.Remove(delDir); err != nil{
		return err
	}
	log.Info("dirs cleaned")
	return nil
}