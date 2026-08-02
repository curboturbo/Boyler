package imageservice

import (
	"boyler/internal/daemon/core"
	"boyler/internal/daemon/infrastructure/outbound/image"
	overlay "boyler/internal/daemon/infrastructure/outbound/overlay"
	"context"
	"sync"
)


type Stream interface{
	Send(*core.PullingEvent) error
}


type ImageSerivceConfig struct {
	UnpackDir string
}


type ImageService interface {
	Pull(ctx context.Context,name string, stream Stream) error
	Remove(ctx context.Context, cmd RemoveCommand) error
	List(ctx context.Context) ([]*core.Image, error)
}

type imageService struct {
	fs      overlay.VolumeManager
	image image.ImageManager
	config ImageSerivceConfig
}

func NewImageService(config ImageSerivceConfig, im image.ImageManager, fs overlay.VolumeManager) ImageService {
	return &imageService{config: config, image: im, fs: fs}
}


func (p *imageService) Pull(ctx context.Context, name string, stream Stream) error {
	ctx, cancel := context.WithCancel(ctx)
	ch := make(chan *core.PullingEvent, 150)
	defer cancel()
	var wg sync.WaitGroup
	wg.Add(1)
	go sendToStream(ctx, stream, ch, &wg)
	if err := p.image.Pull(ctx,name,ch); err != nil {
		wg.Wait()
		return &core.ImageError{Image:name,Err:err}
	}
	wg.Wait()
	return nil
}


func sendToStream(ctx context.Context, stream Stream, ch <-chan *core.PullingEvent, wg *sync.WaitGroup) {
	defer wg.Done()
	for {
		select {
		case <-ctx.Done():
			return
		case val, ok := <-ch:
			if !ok {
				return
			}
			if err := stream.Send(val); err != nil{
				return
			}
		}
	}
}

func (p *imageService) Remove(ctx context.Context, cmd RemoveCommand) error{
	return nil
}
	
func (p *imageService) List(ctx context.Context) ([]*core.Image, error){
	return []*core.Image{}, nil
}