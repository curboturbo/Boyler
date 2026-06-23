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

func NewOverlayManager(baseDir string, containerDir string) VolumeManager{
	return &overlayManager{baseDir: baseDir, containerDir: containerDir}
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


	opts := fmt.Sprintf("lowerdir=%s,upperdir=%s,workdir=%s", lowerDir, upperDir, workDir)
	err := syscall.Mount("overlay", mer)

}