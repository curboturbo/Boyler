package application

import (
	net "boyler/internal/daemon/application/network_service"
	layer "boyler/internal/daemon/infrastructure/outbound/image"
	overlay "boyler/internal/daemon/infrastructure/outbound/overlay"
	registry "boyler/internal/daemon/infrastructure/outbound/registry"
	storage "boyler/internal/daemon/infrastructure/outbound/storage/in-memory"
	limits "boyler/internal/daemon/infrastructure/outbound/limits"
	run "boyler/internal/runtime"
	"log/slog"
)


type Deps struct {
	Runtime run.Runtime
	FS      overlay.VolumeManager
	Images  layer.ImageManager
	Network net.NetworkService
	Reg     registry.ResourcesRegistry
	Store   *storage.ContainerRepository
	Logger  *slog.Logger
	Conf    ServiceConfig
	CgroupFactory limits.Factory
}

type ServiceConfig struct {
	UnpackDir    string
	ContainerDir string
	CgroupPath   string
	SystemPath   string
}
