package application

import (
	core "boyler/internal/daemon/core"
	limits "boyler/internal/daemon/infrastructure/outbound/limits"
	layer "boyler/internal/daemon/infrastructure/outbound/image"
	overlay "boyler/internal/daemon/infrastructure/outbound/overlay"
	net "boyler/internal/daemon/application/network_service"
	registry "boyler/internal/daemon/infrastructure/outbound/registry"
	storage "boyler/internal/daemon/infrastructure/outbound/storage/in-memory"
	run "boyler/internal/runtime"
	"context"
	"log/slog"
	"path/filepath"
	"time"
)

type Creator struct {
	logger  *slog.Logger
	runtime run.Runtime
	fs      overlay.VolumeManager
	images  layer.ImageManager
	network net.NetworkService
	reg     registry.ResourcesRegistry
	store   *storage.ContainerRepository
	conf    ServiceConfig
	cgroupFactory limits.Factory
}

func NewCreator(d Deps) *Creator {
	return &Creator{
		logger:  d.Logger,
		runtime: d.Runtime,
		fs:      d.FS,
		images:  d.Images,
		network: d.Network,
		reg:     d.Reg,
		store:   d.Store,
		cgroupFactory: d.CgroupFactory,
		conf:    d.Conf,
	}
}


func (c *Creator) ExecuteCreate(ctx context.Context, cmd CreateContainerCommand) (*CreateContainerResponse, error) {
	// id = uuid.New().String()
	id := "8e6ff240-a1a5-4642-a58e-1a53b3222ee0" // using by default like test

	if err := c.extractImage(ctx, "alpine"); err != nil {
		return nil, err
	}

	state, pid, createTime, startTime, err := c.provision(ctx, id, cmd.ImageName, &cmd.Limits)
	if err != nil {
		return nil, err
	}

	containerCore := MapApplicationToCore(
		&cmd,
		WithId(id),
		WithPid(int64(pid)),
		WithTime(createTime, startTime),
	)
	c.store.Save(ctx, *containerCore)

	return &CreateContainerResponse{
		ContainerContext: ContainerContext{ID: id},
		Status:           string(state.Status),
	}, nil
}

func (c *Creator) ExecuteStart(ctx context.Context, cmd StartContainerCommand) (*StartContainerResponse, error) {
	container, err := c.store.Get(ctx, cmd.ID)
	if err != nil {
		return nil, &core.InvalidUserCommandError{Op: "start", Err: err}
	}

	id := cmd.ID
	_, pid, _, startTime, err := c.provision(ctx, id, container.ImageID, &container.Config.Resources)
	if err != nil {
		return nil, err
	}

	container.StartedAt = startTime
	container.PID = pid
	c.store.Save(ctx, *container)

	return &StartContainerResponse{
		ContainerContext: ContainerContext{ID: id},
		PID:              int64(pid),
	}, nil
}


func (c *Creator) provision(ctx context.Context, id, imageName string, lim *core.Restriction) (state *run.State, pid int, createTime, startTime time.Time, err error) {
	if err = c.prepareFilesystem(ctx, id, imageName); err != nil {
		return nil, 0, time.Time{}, time.Time{}, err
	}

	bundlePath := filepath.Join(c.conf.ContainerDir, id)
	createTime = time.Now()
	if err = c.createRuntimeContainer(ctx, id, bundlePath); err != nil {
		return nil, 0, time.Time{}, time.Time{}, err
	}

	state, pid, err = c.getContainerState(ctx, id)
	if err != nil {
		return nil, 0, time.Time{}, time.Time{}, err
	}

	if err = c.setupNetwork(ctx, id, pid); err != nil {
		return nil, 0, time.Time{}, time.Time{}, err
	}

	if err = c.applyCgroupLimits(ctx, id, pid, lim); err != nil {
		return nil, 0, time.Time{}, time.Time{}, err
	}

	startTime = time.Now()
	if err = c.startRuntime(ctx, id); err != nil {
		return nil, 0, time.Time{}, time.Time{}, err
	}

	return state, pid, createTime, startTime, nil
}

func (c *Creator) extractImage(ctx context.Context, image string) error {
	if err := c.images.Extract(ctx, image, c.conf.UnpackDir); err != nil {
		c.logger.Error("Failed to extract container image",
			"err", err, "image", image, "unpack_dir", c.conf.UnpackDir)
		return &core.ImageError{Image: image, Err: err}
	}
	return nil
}

func (c *Creator) prepareFilesystem(ctx context.Context, id, imageName string) error {
	if err := c.fs.CreateMountPoints(ctx, id); err != nil {
		c.logger.Error("Failed to prepare container filesystem", "err", err, "container_id", id)
		return &core.FilesystemError{Op: "create_mount_points", Err: err}
	}

	if err := c.fs.Mount(ctx, id, imageName); err != nil {
		c.logger.Error("Failed to mount container filesystem", "err", err, "image_name", imageName)
		return &core.FilesystemError{Op: "mount", Err: err}
	}

	return nil
}

func (c *Creator) createRuntimeContainer(ctx context.Context, id, bundlePath string) error {
	if err := c.runtime.Create(ctx, id, bundlePath); err != nil {
		c.logger.Error("Failed to create container in runtime", "err", err, "bundle_path", bundlePath)
		return &core.RuntimeError{Op: "create", Err: err}
	}
	return nil
}

func (c *Creator) getContainerState(ctx context.Context, id string) (*run.State, int, error) {
	state, err := c.runtime.State(ctx, id)
	if err != nil {
		c.logger.Error("Failed to get state of created container", "err", err, "container_id", id)
		return nil, 0, &core.RuntimeError{Op: "state", Err: err}
	}
	return state, state.PID, nil
}

func (c *Creator) setupNetwork(ctx context.Context, id string, pid int) error {
	if err := c.network.ConnectContainer(ctx, id, pid); err != nil {
		c.logger.Error("Failed to connect container", "err", err, "container_id", id, "pid", pid)
		return &core.NetworkError{Op: "connect", Err: err}
	}
	return nil
}

func (c *Creator) applyCgroupLimits(ctx context.Context, id string, pid int, lim *core.Restriction) error {
	cgroupsManager, err := c.cgroupFactory.New(c.conf.CgroupPath, id, lim, c.conf.SystemPath)
	if err != nil {
		c.logger.Error("Failed to create cgroup-manager", "err", err)
		return &core.CgroupsError{Op: "create_manager", Err: err}
	}

	if err := cgroupsManager.Apply(ctx, uint64(pid)); err != nil {
		c.logger.ErrorContext(ctx, "Failed to apply cgroups limits",
			slog.String("container_id", id),
			slog.Int("pid", pid),
			slog.String("error", err.Error()),
		)
		return &core.CgroupsError{Op: "apply", Err: err}
	}

	if err := c.reg.Post(id, cgroupsManager); err != nil {
		c.logger.Error("Failed to register cgroups manager", "err", err, "container_id", id)
		return &core.CgroupsError{Op: "register", Err: err}
	}

	return nil
}

func (c *Creator) startRuntime(ctx context.Context, id string) error {
	if err := c.runtime.Run(ctx, id); err != nil {
		c.logger.Error("Failed to start container", "err", err, "container_id", id)
		return &core.RuntimeError{Op: "run", Err: err}
	}
	return nil
}