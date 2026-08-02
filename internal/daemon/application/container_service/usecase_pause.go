package application

import (
	"boyler/internal/daemon/core"
	registry "boyler/internal/daemon/infrastructure/outbound/registry"
	storage "boyler/internal/daemon/infrastructure/outbound/storage/in-memory"
	"boyler/pkg/logger"
	"context"
)

type Pauser struct{
	reg     registry.ResourcesRegistry
	store   *storage.ContainerRepository
}

func NewPauser(d Deps) *Pauser {
	return &Pauser{
		reg: d.Reg,
		store: d.Store,
	}
}

func (p *Pauser) Execute(ctx context.Context, cmd PauseContainerCommand) (*PauseContainerResponse, error) {
	ctx = logger.WithFields(ctx, "id", cmd.ID)
	err := p.pauseSignalToCgroup(ctx, cmd.ID)
	if err != nil{ return nil, err}
	container, err := p.store.Get(ctx, cmd.ID)
	container.Status = core.StatusFreeze
	if err = p.store.Update(ctx, *container); err != nil{
		return nil, &core.InternalDaemonError{Op:"update", Err: err}
	}
	return &PauseContainerResponse{
		ContainerContext: ContainerContext{ID: cmd.ID},
	}, nil	
}


func (p *Pauser) pauseSignalToCgroup(ctx context.Context, id string) error {
	manager, err := p.reg.Get(id)
	if err != nil{
		return &core.CgroupsError{Op:"pause", Err: err}
	}
	if err = manager.Freeze(ctx); err != nil {
		return &core.CgroupsError{Op:"freeze", Err: err}
	}
	return nil
}
