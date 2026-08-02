package application

import (
	net "boyler/internal/daemon/application/network_service"
	core "boyler/internal/daemon/core"
	layer "boyler/internal/daemon/infrastructure/outbound/image"
	overlay "boyler/internal/daemon/infrastructure/outbound/overlay"
	registry "boyler/internal/daemon/infrastructure/outbound/registry"
	storage "boyler/internal/daemon/infrastructure/outbound/storage/in-memory"
	run "boyler/internal/runtime"
	"boyler/pkg/logger"
	"context"
	"syscall"
	"time"
)

type Stopper struct{
	runtime run.Runtime
	fs      overlay.VolumeManager
	images  layer.ImageManager
	network net.NetworkService
	reg     registry.ResourcesRegistry
	store   *storage.ContainerRepository
	conf    ServiceConfig
}

func NewStopper(d Deps) *Stopper {
	return &Stopper{
		runtime: d.Runtime,
		fs:      d.FS,
		images:  d.Images,
		network: d.Network,
		reg:     d.Reg,
		store:   d.Store,
		conf:    d.Conf,
	}
}

func (s *Stopper) Execute(ctx context.Context, cmd StopContainerCommand) (*StopContainerResponse, error) {
	id := cmd.ID
	container, err := s.store.Get(ctx, id)
	if err != nil {
		return nil, &core.InternalDaemonError{Op:"core container", Err: err}
	}
	if container.Status == core.StatusRunning {
		if err := s.killMasterProcess(ctx, id); err != nil {
			return nil, err
		}

		if err := s.killCgroup(ctx, id); err != nil {
			return nil, err
		}

		if err := s.unmountFilesystem(ctx, id); err != nil {
			return nil, err
		}
		container.Status = core.StatusStopped
		s.store.Update(ctx, *container)
		if err := s.network.FreeIpAddress(ctx, cmd.ID); err != nil {
			return nil, err
		}
		return &StopContainerResponse{
			ContainerContext:ContainerContext{ID: id},
			}, nil
	}else{
		return nil, &core.InvalidUserCommandError{Op: "stop", Err: core.ErrContainerNotRunning}
	}
}

func (s *Stopper) killMasterProcess(ctx context.Context, id string) error {
	runcCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
	log := logger.FromContext(ctx)
	defer cancel()
	err := s.runtime.Kill(runcCtx, id, syscall.SIGTERM)
	if err != nil {
		log.Warn("Failed to SIGTERM process, send SIGKILL")
		warnCtx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		err = s.runtime.Kill(warnCtx, id, syscall.SIGKILL)
		if err != nil {
			log.Error("Failed to kill process, memory leak, fix", "err", err)
			return &core.RuntimeError{Op: "kill", Err: err}
		}
		return nil
	}
	return nil
}

func (s *Stopper) killCgroup(ctx context.Context, id string) error {
	managerCgroup, err := s.reg.Get(id)
	if err != nil {
		return &core.CgroupsError{Op: "container does not exist", Err: err}
	}
	containerCore, err := s.store.Get(ctx, id)
	if err != nil {
		return &core.InternalDaemonError{Op: "cgroup", Err: err}
	}

	if err := managerCgroup.Delete(ctx, uint64(containerCore.PID)); err != nil {
		return &core.CgroupsError{Op: "cgroup", Err: err}
	}
	return nil
}

func (s *Stopper) unmountFilesystem(ctx context.Context, id string) error {
	if err := s.fs.Unmount(ctx, id); err != nil{
		return &core.FilesystemError{Op:"save fs", Err: err}
	}
	return nil
}