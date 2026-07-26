package application

import (
	"context"
	"fmt"
	"log/slog"
	"path/filepath"

	net "boyler/internal/daemon/application/network_service"
	// core "boyler/internal/daemon/core"
	layer "boyler/internal/daemon/infrastructure/outbound/image"
	limits "boyler/internal/daemon/infrastructure/outbound/limits"
	overlay "boyler/internal/daemon/infrastructure/outbound/overlay"
	registry "boyler/internal/daemon/infrastructure/outbound/registry"
	storage "boyler/internal/daemon/infrastructure/outbound/storage/in-memory"
	run "boyler/internal/runtime"
	logger "boyler/pkg/logger"
	//"github.com/google/uuid"
)


type ContainerService interface {
	CreateAndStart(ctx context.Context, cmd CreateContainerCommand) (*CreateContainerResponse, error)
	Start(ctx context.Context, cmd StartContainerCommand) (*StartContainerResponse, error)
	Stop(ctx context.Context, cmd StopContainerCommand) (*StopContainerResponse, error)
	Delete(ctx context.Context, cmd DeleteContainerCommand) (*DeleteContainerResponse, error)
	Restart(ctx context.Context, cmd RestartContainerCommand) (*RestartContainerResponse, error)
	Attach(ctx context.Context, cmd AttachContainerCommand) (*AttachSession, error) // most harders part all grpc
}

type containerService struct {
	logger *slog.Logger

	runtime run.Runtime

	fs     overlay.VolumeManager
	images layer.ImageManager

	network net.NetworkService
	reg     registry.ResourcesRegistry

	store *storage.ContainerRepository

    conf ServiceConfig
}

type ServiceConfig struct {
    UnpackDir string
    ContainerDir string
    CgroupPath string
    SystemPath string
}

func NewContainerService(
	runtime run.Runtime,
	fs overlay.VolumeManager,
	images layer.ImageManager,
	network net.NetworkService,
    conf ServiceConfig,
) ContainerService {
	return &containerService{
		runtime: runtime,
		fs:      fs,
		images:  images,
		network: network,
		reg:     registry.NewRepo(),
		logger:  logger.InitLogger(false),
		store: storage.NewContainerRepository(),
        conf:conf,
	}
}

func (c *containerService) CreateAndStart(ctx context.Context, cmd CreateContainerCommand) (*CreateContainerResponse, error) {
    // id = uuid.New().String()
    id := "8e6ff240-a1a5-4642-a58e-1a53b3222ee0"

    if err := c.images.Extract(ctx, "alpine", c.conf.UnpackDir); err != nil {
        c.logger.Error("Failed to extract container image", "err", err, "image", "alpine", "unpack_dir", c.conf.UnpackDir)
        return nil, err
    }

    if err := c.fs.CreateMountPoints(ctx, id); err != nil {
        c.logger.Error("Failed to prepare container filesystem", "err", err)
        return nil, fmt.Errorf("Failed prepare filesystem container: %v", err)
    }
    if err := c.fs.Mount(ctx, id, cmd.ImageName); err != nil {
        c.logger.Error("Failed to mount container filesystem", "err", err, "image_name", cmd.ImageName)
        return nil, fmt.Errorf("Failed mount filesystem container: %v", err)
    }

    bundlePath := filepath.Join(c.conf.ContainerDir, id)

    err := c.runtime.Create(ctx, id, bundlePath)
    if err != nil {
        c.logger.Error("Failed to create container in runtime", "err", err, "bundle_path", bundlePath)
        return nil, fmt.Errorf("Failed create container: %v", err)
    }

    state, err := c.runtime.State(ctx, id)
    if err != nil{
        c.logger.Error("Failed find state created container in runc", "err", err)
    }
    pid := state.PID
    
    err = c.network.InitHostNetwork(ctx)
    if err != nil {
        c.logger.Error("Failed to init host network", "err", err)
        return nil, fmt.Errorf("Bridge error: %v", err)
    }

    if err = c.network.ConnectContainer(ctx, id, pid); err != nil {
        c.logger.Error("Failed to connect container", "err", err)
        return nil, fmt.Errorf("Failed connect container error: %v", err)
    }

    cgroupsManager, err := limits.NewResourcesManager(
        c.conf.CgroupPath,
        id,
        &cmd.Limits,
        c.conf.SystemPath,
    )

    if err != nil {
        c.logger.Error("Failed to create cgroup-manager", "err", err)
        return nil, fmt.Errorf("Failed set up cgroups: %w", err)
    }

    if err := cgroupsManager.Apply(ctx, uint64(pid)); err != nil {
        c.logger.ErrorContext(ctx, "Failed to apply cgroups limits",
            slog.String("container_id", id),
            slog.Int("pid", pid),
            slog.String("error", err.Error()),
        )
        return nil, fmt.Errorf("Failed apply cgroups settings: %w", err)
    }
    if err = c.reg.Post(id, cgroupsManager); err != nil{
        return nil, err
    }
    err = c.runtime.Run(ctx, id)
    if err != nil {
        return nil, fmt.Errorf("Failed start container: %v", err)
    }
    return &CreateContainerResponse{
        ContainerContext: ContainerContext{ID: id},
        Status:           string(state.Status),
    }, nil
}




func (c *containerService) Start(ctx context.Context, cmd StartContainerCommand) (*StartContainerResponse, error){
	return nil, nil
}














func (c *containerService) Stop(ctx context.Context, cmd StopContainerCommand) (*StopContainerResponse, error){
	return nil, nil
}
















func (c *containerService) Delete(ctx context.Context, cmd DeleteContainerCommand) (*DeleteContainerResponse, error){
	return nil, nil
}
	
















func (c *containerService) Restart(ctx context.Context, cmd RestartContainerCommand) (*RestartContainerResponse, error) {
	return nil, nil
}




















func (c *containerService) Attach(ctx context.Context, cmd AttachContainerCommand) (*AttachSession, error) {
	return nil, nil
}