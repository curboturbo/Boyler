package cri

import (
	im "boyler/internal/daemon/image"
	overlay "boyler/internal/daemon/overlay"
	runtime "boyler/internal/runtime/myrunc"
	"context"
	"log/slog"
	"os"
	"path/filepath"
	"fmt"
)

func Run() {
	// 1. Настраиваем JSON-логгер
	handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	})
	logger := slog.New(handler)

	logger.Info("Starting Boyler container daemon initialization...")

	// 2. Автоматически определяем корень проекта ~/Boyler
	wd, err := os.Getwd()
	if err != nil {
		logger.Error("Failed to get working directory", "error", err)
		return
	}

	projectRoot := wd
	if filepath.Base(wd) == "bin" {
		projectRoot = filepath.Dir(wd)
	}

	// 3. Формируем абсолютные пути внутри Boyler/lib
	os.Getenv()
	imagesPath := filepath.Join(projectRoot, "lib", "images")         // ~/Boyler/lib/images
	containersPath := filepath.Join(projectRoot, "lib", "containers") // ~/Boyler/lib/containers
	runtimeBinPath := filepath.Join(projectRoot, "bin/myrunc")               // ~/Boyler/bin

	logger.Info("Resolved absolute paths for development",
		"project_root", projectRoot,
		"images_dir", imagesPath,
		"containers_dir", containersPath,
	)

	// 4. Инициализируем ImageManager
	imageManager := im.NewImageManager(imagesPath, logger)
	
	// 🔥 ФИКС ТУТ: Распаковываем архив в кэш образов (в rootfs), а не в папку контейнера
	imageRootfsDir := filepath.Join(imagesPath, "alpine", "rootfs") 

	logger.Info("Extracting rootfs archive into images cache...", "image", "alpine", "target_dir", imageRootfsDir)
	if err := imageManager.Extract("alpine", imageRootfsDir); err != nil {
		logger.Error("Failed to extract alpine image", "error", err)
		return
	}

	// 5. Инициализируем OverlayManager и монтируем слои для контейнера "a"
	containerID := "a"
	logger.Info("Preparing OverlayFS mount points...", "container_id", containerID)
	overManager := overlay.NewOverlayManager(imagesPath, containersPath, logger)
	
	if err := overManager.CreateMountPoints(containerID); err != nil {
		logger.Error("Failed to create overlay directories", "error", err)
		return
	}
	
	logger.Info("Mounting OverlayFS layers...", "container_id", containerID)
	if err := overManager.Mount(containerID, "alpine"); err != nil {
		logger.Error("Failed to mount overlayfs", "error", err)
		return
	}

	// 6. Вызываем твой кастомный runc (myrunc)
	logger.Info("Invoking low-level runtime to create container...", "runtime", "myrunc")
	manager := runtime.NewMyRunc(runtimeBinPath)
	
	// Папка бандла контейнера, где myrunc будет искать config.json и rootfs (merged)
	unpackDir := filepath.Join(containersPath, containerID) 
	if _, err := manager.Create(context.TODO(), containerID, unpackDir); err != nil {
		logger.Error("Runtime creation failed", "error", err)
		return
	}

	logger.Info("Container successfully initialized by runtime!")
}

func Check(){
	dir, err := os.Getwd()
	if err != nil {
		fmt.Println("Ошибка:", err)
		return
	}
	fmt.Println("Текущая директория:", dir)
}