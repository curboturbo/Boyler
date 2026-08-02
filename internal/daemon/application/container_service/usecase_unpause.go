package application

import (
	"boyler/internal/daemon/core"
	registry "boyler/internal/daemon/infrastructure/outbound/registry"
	storage "boyler/internal/daemon/infrastructure/outbound/storage/in-memory"
	"boyler/pkg/logger"
	"context"
)

type Unpauser struct{
	reg     registry.ResourcesRegistry
	store   *storage.ContainerRepository
}

func NewUnpauser(d Deps) *Unpauser {
	return &Unpauser{
		reg: d.Reg,
		store: d.Store,
	}
}

func (u *Unpauser) Execute(ctx context.Context, cmd UnpauseContainerCommand) (*UnpauseContainerResponse, error) {
	ctx = logger.WithFields(ctx, "id", cmd.ID)
	if err := u.pauseSignalToCgroup(ctx, cmd.ID); err != nil{
		return nil, err
	}
	return &UnpauseContainerResponse{
		ContainerContext: ContainerContext{ID: cmd.ID},
	}, nil
}


func (u *Unpauser) pauseSignalToCgroup(ctx context.Context, id string) error {
	manager, err := u.reg.Get(id)
	if err != nil{
		return &core.CgroupsError{Op:"pause", Err: err}
	}
	if err := manager.Unfreeze(ctx); err != nil {
		return &core.CgroupsError{Op:"unfreeze", Err: err}
	}
	return nil
}