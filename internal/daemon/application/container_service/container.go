package application

import (
	"context"
	"fmt"

	core "boyler/internal/daemon/core"
	layer "boyler/internal/daemon/infrastructure/outbound/image"
	"boyler/internal/daemon/infrastructure/outbound/network"
	net "boyler/internal/daemon/infrastructure/outbound/network"
	overlay "boyler/internal/daemon/infrastructure/outbound/overlay"
	run "boyler/internal/runtime"

	"github.com/google/uuid"
)


type ContainerService interface {
	CreateAndStart(ctx context.Context, container core.Container, network core.Net) (*run.State, error)
	Start(ctx context.Context, conatainer core.Container) error
	Stop(ctx context.Context, container core.Container) error
	Delete(ctx context.Context, container core.Container) error
}


type containerService struct{
	runtime run.Runtime
	fs overlay.VolumeManager
	images layer.ImageManager
	network net.NetworkManager

}


func NewContainerService(
	runtime run.Runtime, 
	fs overlay.VolumeManager, 
	images layer.ImageManager,
	network net.NetworkManager,
	) ContainerService{
	return &containerService{
		runtime: runtime,
		fs : fs,
		images: images,
		network: network,
	}
}


func (c *containerService) CreateAndStart(ctx context.Context, container core.Container, bundlePath string) (*run.State, error) {
	container.ID = uuid.New().String()
	if err = c.fs.CreateMountPoints(container.ID); err != nil {
		return &run.State{}, fmt.Errorf("Failed prepare filesystem container: %v", err)
	}
	if err = c.fs.Mount(container.ID,container.ImageID); err != nil {
		return &run.State{}, fmt.Errorf("Failed mount filesystem container: %v", err)
	}
	state, err := c.runtime.Create(ctx, container.ID,bundlePath)
	if err != nil{
		return &run.State{}, fmt.Errorf("Failed create container: %v", err)
	}

	c.




	if err = c.runtime.Run(ctx,state.ID); err != nil{
		return &run.State{}, fmt.Errorf("Failed start container: %v", err)
	}
	return



	return &run.State{}, nil
}




