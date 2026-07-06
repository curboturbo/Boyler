package application

import (
	core "boyler/internal/daemon/core"
	layer "boyler/internal/daemon/infrastructure/outbound/image"
	net "boyler/internal/daemon/application/network_service"
	overlay "boyler/internal/daemon/infrastructure/outbound/overlay"
	registry "boyler/internal/daemon/infrastructure/outbound/registry"
	run "boyler/internal/runtime"
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/google/uuid"
)


type ContainerService interface {
	CreateAndStart(ctx context.Context, container core.Container, bundlePath string) error
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


func (c *containerService) CreateAndStart(ctx context.Context, container core.Container, bundlePath string) error {
	container.ID = uuid.New().String()
	if err := c.fs.CreateMountPoints(container.ID); err != nil {
		return fmt.Errorf("Failed prepare filesystem container: %v", err)
	}
	if err := c.fs.Mount(container.ID,container.ImageID); err != nil {
		return fmt.Errorf("Failed mount filesystem container: %v", err)
	}
	state, err := c.runtime.Create(ctx, container.ID, bundlePath)
	if err != nil{
		return fmt.Errorf("Failed create container: %v", err)
	}
	
	if err = c.runtime.Run(ctx,state.ID); err != nil{
		return fmt.Errorf("Failed start container: %v", err)
	}
	return nil
}
