package networkservice

import (
	net "boyler/internal/daemon/infrastructure/outbound/network"
	"fmt"
	stnet "net"
	"sync"
)

type NetworkServiceConfig struct {
	BridgeName string
	BridgeIP string
	InternalNetwork string // x.x.x.x/24
}

type NetworkService interface{
	InitHostNetwork() error
	ConnectContainer(containerID string, containerPID int) error
	ExposePort(hostPort string, containerPort string, containerIP string) error
	IsolateContainer(containerIP string) error
	UnisolateContainer(containerIP string) error
}


type ContainerNetInfo struct {
	IpAddres string
	Veth string
}


type networkService struct {
	mu             sync.Mutex
	manager net.NetworkManager
	config         NetworkServiceConfig
	containerNet map[string]ContainerNetInfo
	openName []byte
}


func NewNetworkService(manager net.NetworkManager, config NetworkServiceConfig) NetworkService {
	return &networkService{
		manager: manager,
		config: config,
		containerNet: make(map[string]ContainerNetInfo,10),
		openName: make([]byte, 10),
	}
}

func (ns *networkService) InitHostNetwork() error {
	if err := ns.manager.SetUpBridge(ns.config.BridgeName,ns.config.BridgeIP); err != nil{return err}
	if err := ns.manager.BridgeOpen(ns.config.InternalNetwork, ns.config.BridgeName); err !=nil{return err}
	return nil
}

func (ns *networkService) ConnectContainer(containerID string, containerPID int) error {
	vethName := "v" + containerID[:5]
	num, err := ns.AllocateIP()
	if err != nil {return err}

	containerIP, err := hostCIDR(ns.config.InternalNetwork, byte(num))
	if err != nil {return err}

	err = ns.manager.CreateVethPair(containerPID,vethName)
	if err != nil {return err}

	if err = ns.manager.BindVethToBridge(vethName,ns.config.BridgeName); err != nil{return err}

	err = ns.manager.SetupContainerNetwork(containerPID, vethName, containerIP, ns.config.BridgeName, ns.config.BridgeIP)
	if err != nil{return err}

	ns.mu.Lock()
	defer ns.mu.Unlock()
	ns.containerNet[containerID] = ContainerNetInfo{Veth:vethName,IpAddres: containerIP}
	return nil
}

func (ns *networkService) IsolateContainer(containerIP string) error {
	if check := isValidIPv4(containerIP); check == false{
		return fmt.Errorf("invalid containerIP address")
	}
	if err := ns.manager.CreateIsolation(containerIP, ns.config.BridgeName); err != nil{
		return err
	}
	return nil
}


func (ns *networkService) UnisolateContainer(containerIP string) error {
	if check := isValidIPv4(containerIP); check == false{
		return fmt.Errorf("invalid containerIP address")
	}
	if err := ns.manager.RestoreIsolation(containerIP, ns.config.BridgeName); err != nil{
		return err
	}
	return nil
}


func (ns *networkService) ExposePort(hostPort string, containerPort string, containerIP string) error {
	if check := isValidIPv4(containerIP); check == false{
		return fmt.Errorf("invalid containerIP address")
	}
	err := ns.manager.ForwardPort(hostPort, containerPort,containerIP,ns.config.BridgeName)
	if err != nil{
		return err
	}
	return nil
}


func (ns *networkService) AllocateIP() (int, error) {
	ns.mu.Lock()
	defer ns.mu.Unlock()
	ns.openName[0], ns.openName[1] = 1,1
	for i := range ns.openName{
		if ns.openName[i] == 0{
			ns.openName[i] = 1
			return i, nil
		}
	}
	return -1, fmt.Errorf("All ip in internal network is busy")
}

func hostCIDR(network string, host byte) (string, error) {
    ip, ipNet, err := stnet.ParseCIDR(network)
    if err != nil {
        return "", err
    }
    ip = ip.To4()
    if ip == nil {
        return "", fmt.Errorf("only IPv4 is supported")
    }
    ip = append(stnet.IP(nil), ip...)
    ip[3] = host
    ones, _ := ipNet.Mask.Size()
    return fmt.Sprintf("%s/%d", ip.String(), ones), nil
}

func isValidIPv4(ip string) bool {
    parsedIP := stnet.ParseIP(ip)
    if parsedIP == nil {
        return false
    }
    return parsedIP.To4() != nil
}