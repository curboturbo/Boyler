package adapters

import (
	"boyler/internal/ports"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"syscall"
)

type overlayManager struct {
	imageDir string // images directory
	containerDir string // containers directory
	logger *slog.Logger
}

func NewOverlayManager(baseDir string, containerDir string) ports.VolumeManager{
	return &overlayManager{baseDir: baseDir, containerDir: containerDir}
}

func (vm *overlayManager) CreateMountPoints(containerID string) error {
	containerPath := filepath.Join(vm.containerDir, containerID)
	err := os.MkdirAll(containerID, 0755)
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
	return nil
}

func (vm *overlayManager) Moub