package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	service "boyler/internal/daemon/application/container_service"
	networkservice "boyler/internal/daemon/application/network_service"
	image "boyler/internal/daemon/infrastructure/outbound/image"
	limits "boyler/internal/daemon/infrastructure/outbound/limits"
	network "boyler/internal/daemon/infrastructure/outbound/network"
	overlay "boyler/internal/daemon/infrastructure/outbound/overlay"
	registry "boyler/internal/daemon/infrastructure/outbound/registry"
	storage "boyler/internal/daemon/infrastructure/outbound/storage/in-memory"
	runtime "boyler/internal/runtime/myrunc"
	"boyler/pkg/logger"
)


type DaemonConfig struct {
	ImagesPath     string
	ContainersPath string
	RuntimeBinPath string

	Network        networkservice.NetworkServiceConfig
	NetworkManager network.Config
	Service        service.ServiceConfig
}

type DaemonFactory struct {
	config DaemonConfig
}

func NewDaemonFactory(config DaemonConfig) *DaemonFactory {
	return &DaemonFactory{config: config}
}

func NewDaemonFactoryFromEnv() *DaemonFactory {
	root := workingDirectory()
	imagesPath := envOr(root, os.Getenv("IMAGE_PATH"))
	containersPath := envOr(root, os.Getenv("CONTAINER_DIR"))

	return NewDaemonFactory(DaemonConfig{
		ImagesPath:     imagesPath,
		ContainersPath: containersPath,
		RuntimeBinPath: envOr(root, os.Getenv("BIN_MYRUNC")),
		NetworkManager: network.Config{
			Eth0:    os.Getenv("DEFAULT_ETH0"),
			Forward: os.Getenv("IP_FORWARDING_PATH"),
		},
		Network: networkservice.NetworkServiceConfig{
			BridgeName:      os.Getenv("BRIDGE_NAME"),
			BridgeIP:        os.Getenv("BRIDGE_IP"),
			InternalNetwork: os.Getenv("CONTAINER_LOCAL_NETWORK"),
		},
		Service: service.ServiceConfig{
			UnpackDir:    envOr(root, os.Getenv("UNPACK_DIR")),
			ContainerDir: containersPath,
			CgroupPath:   os.Getenv("CGROUP_PATH"),
			SystemPath:   os.Getenv("SYSTEM_PATH"),
		},
	})
}


func (d *DaemonFactory) NewDaemon() (service.ContainerService, error) {
	if d == nil {
		return nil, fmt.Errorf("daemon factory is nil")
	}

	networkService, err := networkservice.NewNetworkService(
		network.NewNetworkManager(d.config.NetworkManager),
		d.config.Network,
	)
	if err != nil {
		return nil, fmt.Errorf("create network service: %w", err)
	}

	return service.NewContainerService(service.Deps{
		Runtime:       runtime.NewMyRunc(d.config.RuntimeBinPath),
		FS:            overlay.NewOverlayManager(d.config.ImagesPath, d.config.ContainersPath),
		Images:        image.NewImageManager(d.config.ImagesPath),
		Network:       networkService,
		Reg:           registry.NewRepo(),
		Store:         storage.NewContainerRepository(),
		Logger:        logger.InitLogger(false),
		CgroupFactory: limits.NewFactory(),
		Conf:          d.config.Service,
	}), nil
}

func envOr(node1 string, node2 string) string{ return filepath.Join(node1,node2) }

func workingDirectory() string {
	wd, err := os.Getwd()
	if err != nil {
		return "."
	}
	return wd
}

func projectRoot() (string,error) {
	wd, err := os.Getwd()
	if err != nil {return "", err}
	projectRoot := wd // /home/tema/Boyler
	if filepath.Base(wd) == "bin" {
		projectRoot = filepath.Dir(wd)
	}
	return projectRoot, nil
}