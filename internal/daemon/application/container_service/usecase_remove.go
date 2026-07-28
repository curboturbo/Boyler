package application

import (
	core "boyler/internal/daemon/core"
	net "boyler/internal/daemon/application/network_service"
	overlay "boyler/internal/daemon/infrastructure/outbound/overlay"
	registry "boyler/internal/daemon/infrastructure/outbound/registry"
	storage "boyler/internal/daemon/infrastructure/outbound/storage/in-memory"
	run "boyler/internal/runtime"
	"context"
	"log/slog"
	"syscall"
	"time"
)

type Remover struct {
	logger  *slog.Logger
	runtime run.Runtime
	fs      overlay.VolumeManager
	reg     registry.ResourcesRegistry
	store   *storage.ContainerRepository
	network  net.NetworkService
}

func NewRemover(d Deps) *Remover {
	return &Remover{
		logger:  d.Logger,
		runtime: d.Runtime,
		fs:      d.FS,
		reg:     d.Reg,
		store:   d.Store,
		network: d.Network,
	}
}

func (r *Remover) Execute(ctx context.Context, cmd RemoveContainerCommand) (*RemoveContainerResponse, error) {
	if err := r.killMasterProcess(ctx, cmd.ID); err != nil {
		return nil, err
	}

	if err := r.killCgroup(ctx, cmd.ID); err != nil {
		return nil, err
	}

	if err := r.unmountFilesystemAndClean(ctx, cmd.ID); err != nil {
		return nil, err
	}

	if err := r.network.FreeIpAddress(ctx, cmd.ID); err != nil {
		return nil, err
	}

	return &RemoveContainerResponse{
		ContainerContext: ContainerContext{ID: cmd.ID},
	}, nil
}

func (r *Remover) killMasterProcess(ctx context.Context, id string) error {
	runcCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	err := r.runtime.Kill(runcCtx, id, syscall.SIGTERM)
	if err != nil {
		r.logger.Warn("Failed to SIGTERM process, send SIGKILL")
		warnCtx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		err = r.runtime.Kill(warnCtx, id, syscall.SIGKILL)
		if err != nil {
			r.logger.Error("Failed to kill process, memory leak, fix", "err", err)
			return &core.RuntimeError{Op: "kill", Err: err}
		}
		return nil
	}
	return nil
}

func (r *Remover) killCgroup(ctx context.Context, id string) error {
	managerCgroup, err := r.reg.Get(id)
	if err != nil {
		return &core.CgroupsError{Op: "container does not exist", Err: err}
	}
	containerCore, err := r.store.Get(ctx, id)
	if err != nil {
		return &core.InternalDaemonError{Op: "cgroup", Err: err}
	}

	if err := managerCgroup.Delete(ctx, uint64(containerCore.PID)); err != nil {
		return &core.CgroupsError{Op: "cgroup", Err: err}
	}
	return nil
}

func (r *Remover) unmountFilesystemAndClean(ctx context.Context, id string) error {
	if err := r.fs.Unmount(ctx, id); err != nil {
		return &core.FilesystemError{Op: "fs", Err: err}
	}
	if err := r.fs.Cleanup(ctx, id); err != nil {
		return &core.FilesystemError{Op: "fs", Err: err}
	}
	return nil
}