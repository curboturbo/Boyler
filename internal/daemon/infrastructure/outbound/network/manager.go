package network

import (
	"fmt"
	stdnet "net"
	"os"
	"os/exec"
	"runtime"
	"context"
	"boyler/pkg/logger"
	"github.com/vishvananda/netlink"
	"github.com/vishvananda/netns"
)


type Config struct {
	Eth0 string
	Forward string
}

// the manager who sets up the network for containers
type NetworkInfrastructureManager interface{
	// create virtual network interface
	SetUpBridge(ctx context.Context, bridgeName string, bridgeIp string) error
	// delete bridge
	DeleteBridge(ctx context.Context, bridgeName string) error
	// create virtual veth pair for container and bridge
	CreateVethPair(ctx context.Context, containerPID int, vethName string) error
	// connect veth pair to bridge
	BindVethToBridge(ctx context.Context, vethName string, bridgeName string) error
	// set up nat in host machine
	BridgeOpen(ctx context.Context, internalNetwork string, bridgeName string) error
	// setup all contianer network
	SetupContainerNetwork(ctx context.Context, containerPID int, vethName string, ipAddress string, bridgeName string, bridgeIP string) error
	// show veth
	ShowVeth() []string
	// create isolation (dont delete real L2 connection, for delete veth-pair using)
	CreateIsolation(ctx context.Context, ipAddress string, bridgeName string) error
	// restore isolation
	RestoreIsolation(ctx context.Context, ipAddress string, bridgeName string) error 
	// forward port
	ForwardPort(ctx context.Context, hostPort string, containerPort string,ipAddres string, bridgeName string) error
}


type networkInfrastructureManager struct{
	createdVeths []string
	config Config
	firewall FirewallManager
}


func NewNetworkManager(config Config) NetworkInfrastructureManager{
	createdVeths := []string{}
	firewall := NewFirewallManager()
	return &networkInfrastructureManager{createdVeths: createdVeths, config: config, firewall: firewall}
}


func (net *networkInfrastructureManager) SetUpBridge(ctx context.Context, bridgeName string, bridgeIP string) error {
	log := logger.FromContext(ctx)
	log.Debug("start setup l2/l3 bridge")
	_, err := netlink.LinkByName(bridgeName)
	if err == nil{
		log.Warn("Most has alredy done","bridgeName",bridgeName)
		return nil
	}
	la := netlink.NewLinkAttrs()
	la.Name = bridgeName
	bridge := &netlink.Bridge{LinkAttrs: la}
	if err := netlink.LinkAdd(bridge); err != nil {
		return fmt.Errorf("Failed to create bridge %s: %v", bridgeName, err)
	}
	addr, err := netlink.ParseAddr(bridgeIP)
	if err != nil {
		return fmt.Errorf("Failed invalid bridge IP %s: %v", bridgeIP, err)
	}

	if err := netlink.AddrAdd(bridge, addr); err != nil {
		return fmt.Errorf("Failed to add IP %s to bridge %s: %w", bridgeIP, bridgeName, err)
	}
	log.Debug("setup created bridge")
	if err = netlink.LinkSetUp(bridge); err != nil{
		return fmt.Errorf("Failed to set up bridge: %v", err)
	}
	log.Info("bridge done")
	return nil
}

func (net *networkInfrastructureManager) CreateVethPair(ctx context.Context, contanerPID int, vethName string) error {
	log := logger.FromContext(ctx)
	log.Debug("start creating veth pair for bridge and containers")
	if _, err := netlink.LinkByName(vethName); err == nil {
		return fmt.Errorf("Failed: veth %s already exists", vethName)
	}

	la := netlink.NewLinkAttrs()
	la.Name = vethName
	peerVethName := net.config.Eth0
	veth := &netlink.Veth{LinkAttrs: la, PeerName: peerVethName} 

	if err := netlink.LinkAdd(veth); err != nil {
		return fmt.Errorf("Failed to create veth: %v", err)

	}

	peerLink, err := netlink.LinkByName(peerVethName) 
	if err != nil {
		return fmt.Errorf("Failed to find peer veth %s after creation: %w", peerVethName, err)
	}

	if err = netlink.LinkSetNsPid(peerLink, contanerPID); err != nil {
		if hostLink, errDel := netlink.LinkByName(vethName); errDel == nil {
			netlink.LinkDel(hostLink)
		}
		return fmt.Errorf("Failed to lay second end to PID: %v", err)
	}
	hostLink, err := netlink.LinkByName(vethName)
	if err != nil {
		return fmt.Errorf("Failed to find host veth %s for bring up: %w", vethName, err)
	}
	if err = netlink.LinkSetUp(hostLink); err != nil {
		return fmt.Errorf("Failed to set up host veth %s: %w", vethName, err)
	}
	net.createdVeths = append(net.createdVeths, vethName)
	log.Info("veth-pair done")
	return nil
}


func (net *networkInfrastructureManager) BindVethToBridge(ctx context.Context, vethName string, bridgeName string) error {
	log  := logger.FromContext(ctx)
	log.Debug("start attaching veth and bridge")
	bridge, err := netlink.LinkByName(bridgeName)
	if err != nil {
		return fmt.Errorf("Failed find %v : %v", bridgeName, err)
	}
	hostVeth, err := netlink.LinkByName(vethName)
	if err != nil {
		return fmt.Errorf("Failed find %v : %v", vethName, err)
	}
	if err = netlink.LinkSetMaster(hostVeth, bridge); err != nil{
		return fmt.Errorf("Failed to connect %v to bridge %v : %v", vethName, bridge, err)
	}
	log.Info("binding done")
	return nil
}


func (net *networkInfrastructureManager) SetupContainerNetwork(ctx context.Context, containerPID int, vethName string, ipAddress string, bridgeName string, bridgeIP string) error {
	log := logger.FromContext(ctx)
	log.Debug("start initialization network in container")
	log.Debug("lock all goroutines in thread, for exclude planner error")
    runtime.LockOSThread()
    defer runtime.UnlockOSThread()

    containerNetNamespace, err := netns.GetFromPid(containerPID)
    if err != nil {
        return fmt.Errorf("Failed get container namespace: %v", err)
    }
    defer containerNetNamespace.Close()
	log.Debug("switch thread namespace")

    handle, err := netlink.NewHandleAt(containerNetNamespace)
    if err != nil {
        return fmt.Errorf("Failed to create netlink handle for container: %v", err)
    }
    defer handle.Close()
    peerVethName := net.config.Eth0
    
    link, err := handle.LinkByName(peerVethName)
    if err != nil {
        return fmt.Errorf("Failed to find %s in container: %v", peerVethName, err)
    }

    addr, err := netlink.ParseAddr(ipAddress)
    if err != nil {
        return fmt.Errorf("Failed to parse ip %s : %v", ipAddress, err)
    }

    if err = handle.AddrAdd(link, addr); err != nil {
        return fmt.Errorf("Failed add ip-address to veth-pair: %v", err)
    }

    if err = handle.LinkSetUp(link); err != nil {
        return fmt.Errorf("Failed to set up veth: %v", err)
    }

    if localHost, err := handle.LinkByName("lo"); err == nil {
        handle.LinkSetUp(localHost)
    }

    gatewayIP, _, err := stdnet.ParseCIDR(bridgeIP)
    if err != nil {
        gatewayIP = stdnet.ParseIP(bridgeIP)
        if gatewayIP == nil {
            return fmt.Errorf("invalid bridge IP format for gateway: %s", bridgeIP)
        }
    }
	log.Debug("add route in routing-table")
    route := &netlink.Route{
        Scope:     netlink.SCOPE_UNIVERSE,
        LinkIndex: link.Attrs().Index,
        Gw:        gatewayIP,
    }
    if err = handle.RouteAdd(route); err != nil {
        return fmt.Errorf("Failed to update router rules inside container: %v", err)
    }
	log.Info("network in container done")

    return nil
}


func (net *networkInfrastructureManager) DeleteBridge(ctx context.Context, bridgeName string) error {
	log := logger.FromContext(ctx)
	log.Debug("start delete bridge","bridge_name", bridgeName)
	link, err := netlink.LinkByName(bridgeName)
	if err != nil{
		return fmt.Errorf("Failed to find bridge: %v", err)
	}
	bridgeInd := link.Attrs().Index
	allLink, err := netlink.LinkList()
	if err != nil{
		return fmt.Errorf("Failed to show ip links: %v",err)
	}
	count := 0
	for _, l := range allLink{
		if l.Attrs().MasterIndex == bridgeInd && l.Type() == "veth"{count+=1}
	}
	if count != 0{
		return fmt.Errorf("Failed to delete bridge: attach %d veth",count)
	}else{
		if err = netlink.LinkDel(link); err != nil{
			return fmt.Errorf("Failed to delete bridge: %v", err)
		}
	}
	log.Info("bridge deleted")
	return nil
}

func (net *networkInfrastructureManager) BridgeOpen(ctx context.Context, internalNetwork string, bridgeName string) error {
	log := logger.FromContext(ctx)
	log.Debug("start opening bridge")
	if err := os.WriteFile(net.config.Forward, []byte("1\n"), 0644); err !=nil{
		return fmt.Errorf("Failed enable ip-forwarding: %v",err)
	}
	cmd := exec.Command("iptables", "-t", "nat", "-A", "POSTROUTING", "-s", internalNetwork, "!", "-o", bridgeName, "-j", "MASQUERADE")
    if err := cmd.Run(); err != nil{
		return fmt.Errorf("Failed enable masquerade with host: %v",err)
	}
	log.Info("bridge opened")
	return nil	
}


func (net *networkInfrastructureManager) ShowVeth() []string{
	return net.createdVeths
}


func (net *networkInfrastructureManager) CreateIsolation(ctx context.Context, ipAddress string, bridgeName string) error {
	log := logger.FromContext(ctx)
	log.Info("start container isolation")
	return net.firewall.Isolate(ipAddress, bridgeName)
}

func (net *networkInfrastructureManager) RestoreIsolation(ctx context.Context, ipAddress string, bridgeName string) error {
	log := logger.FromContext(ctx)
	log.Info("start container isolation")
	return net.firewall.CancelIsolation(ipAddress, bridgeName)
}

func (net *networkInfrastructureManager) ForwardPort(ctx context.Context, hostPort string, containerPort string,ipAddres string,bridgeName string) error {
	log := logger.FromContext(ctx)
	log.Info("start container isolation")
	return net.firewall.Forward(hostPort, containerPort, ipAddres,bridgeName)
}