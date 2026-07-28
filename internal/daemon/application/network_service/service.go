package networkservice

import (
	"context"
	"fmt"
	"sync"

	"boyler/internal/daemon/core"
	net "boyler/internal/daemon/infrastructure/outbound/network"
	"boyler/pkg/logger"
)

type NetworkServiceConfig struct {
	BridgeName string
	BridgeIP string  		// x.x.x.x/24
	InternalNetwork string // x.x.x.x/24
}

type NetworkService interface{
	ConnectContainer(ctx context.Context, containerID string, containerPID int) error
	ExposePort(ctx context.Context, hostPort string, containerPort string, containerIP string) error
	IsolateContainer(ctx context.Context, containerIP string) error
	UnisolateContainer(ctx context.Context, containerIP string) error
	FreeIpAddress(ctx context.Context, containerID string) error
}


type ContainerNetInfo struct {
	IpAddres string		// x.x.x.x/24
	Veth string
	IpIndex int
}

// Bridge creation is a singleton operation during the creation of the first container. 
// The current release does not support multi-network management.
// Multi-network management will be available in boiler v1.1.
type networkService struct {
	mu             sync.Mutex
	manager net.NetworkInfrastructureManager
	config         NetworkServiceConfig
	containerNet map[string]ContainerNetInfo
	ipPool []byte
	netBridgeSingleton sync.Once
	hostNetErr *core.NetworkError
}


func NewNetworkService(manager net.NetworkInfrastructureManager, config NetworkServiceConfig) (NetworkService, error) {
	subNetSize, err := UsableHosts(config.InternalNetwork)
	if err != nil || subNetSize == 0{
		return nil, fmt.Errorf("Invalid network")
	}
	return &networkService{
		manager: manager,
		config: config,
		containerNet: make(map[string]ContainerNetInfo,10),
		ipPool: AllocateIpTable(subNetSize),
	}, nil
}


func (ns *networkService) ConnectContainer(ctx context.Context, containerID string, containerPID int) error {
	if err := ns.ensureHostNetwork(ctx); err != nil {
		return &core.NetworkError{Op: "init_host", Err: err}
	}
	ctx = logger.WithFields(ctx,"containerPID", containerPID, "bridge_name",ns.config.BridgeName, "bridge_ip", ns.config.BridgeIP)
	vethName := "v" + containerID[:5]
	num, err := ns.AllocateIP()
	if err != nil {return err}

	containerIP, err := hostCIDR(ns.config.InternalNetwork, byte(num))
	if err != nil {return err}

	err = ns.manager.CreateVethPair(ctx, containerPID,vethName)
	if err != nil {return err}

	if err = ns.manager.BindVethToBridge(ctx,vethName,ns.config.BridgeName); err != nil{return err}

	err = ns.manager.SetupContainerNetwork(ctx, containerPID, vethName, containerIP, ns.config.BridgeName, ns.config.BridgeIP)
	if err != nil{return err}

	ns.mu.Lock()
	defer ns.mu.Unlock()
	ns.containerNet[containerID] = ContainerNetInfo{Veth:vethName,IpAddres: containerIP, IpIndex: num}
	return nil
}

func (ns *networkService) IsolateContainer(ctx context.Context, containerIP string) error {
	ctx = logger.WithFields(ctx, "containerIP", containerIP)
	if check := isValidIPv4(containerIP); check == false{
		return fmt.Errorf("invalid containerIP address")
	}
	if err := ns.manager.CreateIsolation(ctx,containerIP, ns.config.BridgeName); err != nil{
		return err
	}
	return nil
}


func (ns *networkService) UnisolateContainer(ctx context.Context, containerIP string) error {
	ctx = logger.WithFields(ctx, "containerIP", containerIP,"bridge_name", ns.config.BridgeName)
	if check := isValidIPv4(containerIP); check == false{
		return fmt.Errorf("invalid containerIP address")
	}
	if err := ns.manager.RestoreIsolation(ctx, containerIP, ns.config.BridgeName); err != nil{
		return err
	}
	return nil
}


func (ns *networkService) ExposePort(ctx context.Context, hostPort string, containerPort string, containerIP string) error {
	ctx = logger.WithFields(ctx,"host_port",hostPort,"container_port", containerPort, "bridge_name", ns.config.BridgeName)
	if check := isValidIPv4(containerIP); check == false{
		return fmt.Errorf("invalid containerIP address")
	}
	err := ns.manager.ForwardPort(ctx, hostPort, containerPort,containerIP,ns.config.BridgeName)
	if err != nil{
		return err
	}
	return nil
}


func (ns *networkService) AllocateIP() (int, error) {
	ns.mu.Lock()
	defer ns.mu.Unlock()
	for i := range ns.ipPool{
		if ns.ipPool[i] == 0{
			ns.ipPool[i] = 1
			return i, nil
		}
	}
	return -1, fmt.Errorf("All ip in internal network is busy")
}

func (ns *networkService) FreeIp(num int) error{
	ns.mu.Lock()
	defer ns.mu.Unlock()
	if num > len(ns.ipPool) || num <= 1{
		return fmt.Errorf("Invalid destination address")
	}else{ ns.ipPool[num] = 0}
	return nil
}

func (ns *networkService) FreeIpAddress(ctx context.Context, containerID string) error {
	value, ok := ns.containerNet[containerID]
	if !ok {
		return &core.NetworkError{Op:"free", Err: core.ErrIpAddressNotExist}
	}
	err := ns.FreeIp(value.IpIndex)
	if err != nil{
		return &core.NetworkError{Op:"free ip", Err: err}
	}
	delete(ns.containerNet, containerID)
	return nil
}

func (ns *networkService) ensureHostNetwork(ctx context.Context) error {
	ns.netBridgeSingleton.Do(func() {
		ctx = logger.WithFields(ctx, "bridge_name", ns.config.BridgeName, "bridge_ip", ns.config.BridgeIP)
		
		if err := ns.manager.SetUpBridge(ctx, ns.config.BridgeName, ns.config.BridgeIP); err != nil {
			ns.hostNetErr = &core.NetworkError{Op: "bridge", Err: err}
			return
		}
		
		if err := ns.manager.BridgeOpen(ctx, ns.config.InternalNetwork, ns.config.BridgeName); err != nil {
			ns.hostNetErr = &core.NetworkError{Op: "bridge", Err: err}
			return
		}
	})
	if ns.hostNetErr != nil {
		return ns.hostNetErr
	}
	return nil
}