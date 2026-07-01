package overlay

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"syscall"
)

type overlayManager struct {
	imageDir string // images directory {boyler/lib/images}
	containerDir string // containers directory {boyler/lib/containers}
	logger *slog.Logger // logger
}

func NewOverlayManager(imageDir string, containerDir string, logger *slog.Logger) VolumeManager{
	return &overlayManager{
		imageDir: imageDir,
		containerDir: containerDir,
		logger : logger,
	}
}

func (vm *overlayManager) CreateMountPoints(containerID string) error {
	containerPath := filepath.Join(vm.containerDir, containerID)
	err := os.MkdirAll(containerPath, 0755)
	if err != nil {
		vm.logger.Error("error during container creation")
		return err
	}
	merged := filepath.Join(containerPath, "merged")
	work := filepath.Join(containerPath, "work")
	upper := filepath.Join(containerPath,"upper")
	for _, path := range []string{merged, work, upper} {
		err = os.MkdirAll(path,0755)
		if err != nil{
			vm.logger.Error("failed to create directory", slog.String("path", path))
			os.RemoveAll(containerPath)
			return err
		}
	}
	vm.logger.Debug("mount directory was created")
	return nil
}

func (vm *overlayManager) Mount(containerID string, imageName string) error {
	containerPath := filepath.Join(vm.containerDir, containerID)

	mergedDir := filepath.Join(containerPath, "merged")
	upperDir := filepath.Join(containerPath, "upper")
	workDir := filepath.Join(containerPath, "work")
	lowerDir := filepath.Join(vm.imageDir,imageName,"rootfs")

	if _, err := os.Stat(lowerDir); os.IsNotExist(err){
		vm.logger.Error("image does not exist",slog.String("image", imageName))
		return err
	}

	opts := fmt.Sprintf("lowerdir=%s,upperdir=%s,workdir=%s", filepath.Clean(lowerDir), filepath.Clean(upperDir), filepath.Clean(workDir))
	err := syscall.Mount("overlay", mergedDir, "overlay", 0, opts)

	if err != nil {
		vm.logger.Error("failed to mount overlay",slog.String("image",imageName),slog.String("container", containerID))
		return err
	}

	vm.logger.Debug("overlay mounted",
		slog.String("container", containerID),
		slog.String("merged", mergedDir),
	)
	return nil
}


func (vm *overlayManager) Unmount(containerID string) error {
	mergedDir := filepath.Join(vm.containerDir, containerID, "merged")
	return syscall.Unmount(mergedDir, 0)
}


func (vm *overlayManager) Cleanup(containerID string) error {
	delDir := filepath.Join(vm.containerDir, containerID)
	entries, err := os.ReadDir(delDir)
	if err != nil {return err}
	for _, entry := range entries {
		err := os.RemoveAll(filepath.Join(delDir, entry.Name()))
		if err != nil {return err}
	}
	vm.logger.Debug("dirs was cleaned")
	return nil
}