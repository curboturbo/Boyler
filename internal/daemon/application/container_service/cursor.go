package application

import (
	"boyler/internal/daemon/core"
	storage "boyler/internal/daemon/infrastructure/outbound/storage/in-memory"
	"time"
	"context"
)

type Cursor struct{
	store *storage.ContainerRepository
}

func NewCursor(d Deps) *Cursor {
	return &Cursor{
		store: d.Store,
	}
}

func (c *Cursor) Execute(ctx context.Context, cmd InspectContainerCommand) (*InspectContainerResponse, error) {
	return c.findContainer(ctx, cmd.ID)
	
}

func (c *Cursor) findContainer(ctx context.Context, id string) (*InspectContainerResponse, error){
	container, err := c.store.Get(ctx, id)
	if err != nil {
		return nil, &core.InternalDaemonError{Op: "find", Err: err}
	}
	return cursorPrivateMapper(container), nil
}


func (c *Cursor) Ps(ctx context.Context, cmd PsCommand) ([]*core.Container, error) {
	containerList, err  := c.store.List(ctx)
	if err != nil{ return nil, &core.InternalDaemonError{Op:"Ps", Err:err} }
	return containerList, nil
}


func cursorPrivateMapper(container *core.Container) *InspectContainerResponse {
	var maxMem int64
	if container.Config.Resources.Memory.Max != nil {
		maxMem = *container.Config.Resources.Memory.Max
	}

	var cpuWeight uint64
	if container.Config.Resources.CPU.Weight != nil {
		cpuWeight = *container.Config.Resources.CPU.Weight
	}

	var cpuQuota int64
	if container.Config.Resources.CPU.Quota != nil {
		cpuQuota = *container.Config.Resources.CPU.Quota
	}

	var cpuPeriod uint64
	if container.Config.Resources.CPU.Period != nil {
		cpuPeriod = *container.Config.Resources.CPU.Period
	}
	return &InspectContainerResponse{
		ContainerID: container.ID,
		Pid:         int32(container.PID),
		ImageID:     container.ImageID,
		CreatedAt:   container.CreatedAt.Format(time.RFC3339),
		StartedAt:   container.StartedAt.Format(time.RFC3339),
		Status:      string(container.Status),
		Hostname:    container.Config.Hostname,
		Env:         container.Config.Env,
		Args:        container.Config.Args,
		Resources: core.Restriction{
			Memory: core.MemoryRestriction{
				Max:   &maxMem,
			},
			CPU: core.CPURestriction{
				Weight: &cpuWeight,
				Quota:  &cpuQuota,
				Period: &cpuPeriod,
				Cpus:   container.Config.Resources.CPU.Cpus,
				Mems:   container.Config.Resources.CPU.Mems,
			},
		},
	}
}