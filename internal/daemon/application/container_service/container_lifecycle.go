package application

import (
	"context"
	"log/slog"
	limits "boyler/internal/daemon/infrastructure/outbound/limits"
	run "boyler/internal/runtime"
	core "boyler/internal/daemon/core"
)

func (c *containerService) extractImage(ctx context.Context, image string) error {
	if err := c.images.Extract(ctx, image, c.conf.UnpackDir); err != nil {
		c.logger.Error("Failed to extract container image",
			"err", err, "image", image, "unpack_dir", c.conf.UnpackDir)
		return &core.ImageError{Image: image, Err: err}
	}
	return nil
}


func (c *containerService) prepareFilesystem(ctx context.Context, id, imageName string) error {
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


func (c *containerService) createRuntimeContainer(ctx context.Context, id, bundlePath string) error {
	if err := c.runtime.Create(ctx, id, bundlePath); err != nil {
		c.logger.Error("Failed to create container in runtime", "err", err, "bundle_path", bundlePath)
		return &core.RuntimeError{Op: "create", Err: err}
	}
	return nil
}


func (c *containerService) getContainerState(ctx context.Context, id string) (*run.State, int, error) {
	state, err := c.runtime.State(ctx, id)
	if err != nil {
		c.logger.Error("Failed to get state of created container", "err", err, "container_id", id)
		return nil, 0, &core.RuntimeError{Op: "state", Err: err}
	}
	return state, state.PID, nil
}

func (c *containerService) startRuntime(ctx context.Context, id string) error {
	if err := c.runtime.Run(ctx, id); err != nil {
		c.logger.Error("Failed to start container", "err", err, "container_id", id)
		return &core.RuntimeError{Op: "run", Err: err}
	}
	return nil
}


func (c *containerService) setupNetwork(ctx context.Context, id string, pid int) error {
	if err := c.network.InitHostNetwork(ctx); err != nil {
		c.logger.Error("Failed to init host network", "err", err)
		return &core.NetworkError{Op: "init_host", Err: err}
	}

	if err := c.network.ConnectContainer(ctx, id, pid); err != nil {
		c.logger.Error("Failed to connect container", "err", err, "container_id", id, "pid", pid)
		return &core.NetworkError{Op: "connect", Err: err}
	}

	return nil
}


func (c *containerService) applyCgroupLimits(ctx context.Context, id string, pid int, lim *core.Restriction) error {
	cgroupsManager, err := limits.NewResourcesManager(c.conf.CgroupPath, id, lim, c.conf.SystemPath)
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