package application

import (
	net "boyler/internal/daemon/application/network_service"
	core "boyler/internal/daemon/core"
	layer "boyler/internal/daemon/infrastructure/outbound/image"
	overlay "boyler/internal/daemon/infrastructure/outbound/overlay"
	registry "boyler/internal/daemon/infrastructure/outbound/registry"
	run "boyler/internal/runtime"
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	//"github.com/google/uuid"
)


type ContainerService interface {
	CreateAndStart(ctx context.Context, container core.Container) error
	//Start(ctx context.Context, conatainer core.Container) error
	//Stop(ctx context.Context, container core.Container) error
	//Delete(ctx context.Context, container core.Container) error
	//Restart(ctx context.Context, container core.Container) error
}


type containerService struct{
	logger *slog.Logger

	runtime run.Runtime

	fs overlay.VolumeManager
	images layer.ImageManager

	network net.NetworkService
	reg registry.ResourcesRegistry
}


func NewContainerService(
	runtime run.Runtime, 
	fs overlay.VolumeManager, 
	images layer.ImageManager,
	network net.NetworkService,
	reg registry.ResourcesRegistry,
	) ContainerService{
	return &containerService{
		runtime: runtime,
		fs : fs,
		images: images,
		network: network,
		reg: reg,
		logger: slog.New(slog.NewTextHandler(os.Stdout,nil)),
	}
}


func (c *containerService) CreateAndStart(ctx context.Context, container core.Container) error {
	//container.ID = uuid.New().String()
	container.ID = "8e6ff240-a1a5-4642-a58e-1a53b3222ee0"
	container.ImageID = "alpine"

	unpackDir := "/home/tema/Boyler/lib/images/alpine/rootfs"

	c.images.Extract("alpine",unpackDir)

	if err := c.fs.CreateMountPoints(container.ID); err != nil {
		return fmt.Errorf("Failed prepare filesystem container: %v", err)
	}
	c.logger.Debug(fmt.Sprintf("image %v extracted",container.ImageID))
	if err := c.fs.Mount(container.ID, container.ImageID); err != nil {
		return fmt.Errorf("Failed mount filesystem container: %v", err)
	}

	bundlePath := filepath.Join("/home/tema/Boyler/lib/containers", container.ID)

	state, err := c.runtime.Create(ctx, container.ID, bundlePath)
	if err != nil{
		return fmt.Errorf("Failed create container: %v", err)
	}
	fmt.Print(state)
	fmt.Print("\n")

	
	var pid int
	fmt.Scanln(&pid)

	err = c.network.InitHostNetwork()
	if err != nil{
		return fmt.Errorf("Bridge error: %v", err)
	}

	fmt.Printf("Bridge has done\n")
	if err = c.network.ConnectContainer(container.ID, pid); err != nil{
		return fmt.Errorf("Connect container  error: %v", err)
	}
	fmt.Print("Container connected")
	
	state, err = c.runtime.Run(ctx, container.ID)
	if err != nil{
		return fmt.Errorf("Failed start container: %v", err)
	}
	return nil
}
