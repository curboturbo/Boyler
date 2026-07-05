package networkservice

import (
	net "boyler/internal/daemon/infrastructure/outbound/network"
	"fmt"
	"sync"
)

type NetworkServiceConfig struct {
	BridgeName string
	BridgeIP string
	InternalNetwork string
}

type NetworkService interface{
	InitHostNetwork() error
	ConnectContainer(containerID string, containerPID int, containerIP string) (vethName string, err error)
	DisconnectContainer(containerID string) error
	ExposePort(hostPort string, containerPort string, containerIP string) error
	IsolateContainer(containerIP string) error
	UnisolateContainer(containerIP string) error
}

type networkService struct {
	mu             sync.Mutex
	manager net.NetworkManager
	config         NetworkServiceConfig
	containerVeths map[string]string
}


func NewNetworkService(manager net.NetworkManager, config NetworkServiceConfig) NetworkService {
	return &networkService{
		manager: manager,
		config: config,
		containerVeths: make(map[string]string,15),
	}
}

func (ns *networkService) InitHostNetwork() error {
	if err := ns.manager.SetUpBridge(ns.config.BridgeName,ns.config.BridgeIP); err != nil{return err}
	if err := ns.manager.BridgeOpen(ns.config.InternalNetwork, ns.config.BridgeName); err !=nil{return err}
	return nil
}

func (ns *networkService) ConnectContainer(containerID string, containerPID int, containerIP string) (vethName string, err error){
	ns.mu.Lock()
	defer ns.mu.Unlock()
	err := ns.manager.CreateVethPair(containerPID, vethName)
}