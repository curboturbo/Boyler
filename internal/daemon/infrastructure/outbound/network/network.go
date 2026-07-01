package network

import (
	"fmt"
	"os"
	"os/exec"
	stdnet "net"
	"github.com/vishvananda/netlink"
	"github.com/vishvananda/netns"
)

// the manager who sets up the network for containers
type NetwokManager interface{
	// create virtual network interface
	SetUpBridge(bridgeName string, bridgeIp string) error
	// create virtual veth pair for container and bridge
	CreateVethPair(containerPID int, vethName string)
	// connect veth pair to bridge
	BindVethToBridge(vethName string, bridgeName string) error
	// set up nat in host machine
	BridgeOpen(internalNetwork string) error
	// setup all contianer network
	SetupContainerNetwork(containerPID int, vethName string, ipAddress string, bridgeName string, bridgeIP string) error
	// destroy all virtual interfeces
	CleanUp(bridgeName string) error
	// connects the container to an existing network
	AddContainerToBridge(containerPID int)
	// show veth
	ShowVeth() []string
}

// realization Network Manager
type networkManager struct{
	createdVeths []string
}


func (net *networkManager) SetUpBridge(bridgeName string, bridgeIP string) error {
	_, err := netlink.LinkByName(bridgeName)
	if err == nil{
		return fmt.Errorf("Failed bridge has already done: %v", err)
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
	if err = netlink.LinkSetUp(bridge); err != nil{
		return fmt.Errorf("Failed to set up bridge: %v", err)
	}
	return nil
}

func (net *networkManager) CreateVethPair(contanerPID int, vethName string) error {
	if _, err := netlink.LinkByName(vethName); err == nil {
		return fmt.Errorf("Failed: veth %s already exists", vethName)
	}

	la := netlink.NewLinkAttrs()
	la.Name = vethName
	peerVethName := os.Getenv("DEFAULT_ETH0")
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
	return nil
}


func (net *networkManager) BindVethToBridge(vethName string, bridgeName string) error {
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
	return nil
}


// all ip format - x.x.x.x/y
func (net *networkManager) SetupContainerNetwork(containerPID int,vethName string, ipAddress string, bridgeName string, bridgeIP string) error {
	peerVethName := os.Getenv("DEFAULT_ETH0")
	hostNetNamespace, err := netns.Get()
	if err != nil{
		return fmt.Errorf("Failed get host-thread namespaces: %v", err)
	}
	defer hostNetNamespace.Close() // best practices in https://pkg.go.dev/github.com/vishvananda/netns#NsHandle

	containerNetNamespace, err := netns.GetFromPid(containerPID)
	if err != nil{
		return fmt.Errorf("Failed get container-thread namespces: %v", err)
	}
	defer containerNetNamespace.Close()

	if err = netns.Set(containerNetNamespace); err != nil{
		return fmt.Errorf("Failed to set host to container net namespace: %v", err)
	}
	defer netns.Set(hostNetNamespace) // возврат к хостовому сетевому пространтсву

	link, err := netlink.LinkByName(peerVethName)
	if err != nil{
		return fmt.Errorf("Failed to find %s in conrainer: %v", peerVethName, err)
	}
	addr, err := netlink.ParseAddr(ipAddress)
	if err != nil{
		return fmt.Errorf("Failed to parse ip %s : %v",ipAddress, err)
	}
	err = netlink.AddrAdd(link, addr)
	if err != nil{
		return fmt.Errorf("Failed add ip-address to veth-pair: %v", err)
	}
	err = netlink.LinkSetUp(link)
	if err != nil{
		return fmt.Errorf("Failed to set up veth: %v", err)
	}

	localHost, err := netlink.LinkByName("lo")
	if err == nil{
		netlink.LinkSetUp(localHost)
	}

	gatewayIP, _, err := stdnet.ParseCIDR(bridgeIP)
	if err != nil {
		gatewayIP = stdnet.ParseIP(bridgeIP)
		if gatewayIP == nil {
			return fmt.Errorf("invalid bridge IP format for gateway: %s", bridgeIP)
		}
	}

	route := &netlink.Route{
		Scope:     netlink.SCOPE_UNIVERSE,
		LinkIndex: link.Attrs().Index,
		Gw:        gatewayIP,
	}

	if err = netlink.RouteAdd(route); err != nil{
		return fmt.Errorf("Failed to update router s rules: %v", err)
	}

	return nil
}

func (net *networkManager) BridgeOpen(internalNetwork string) error {
	// включаем пересылку 
	if err := os.WriteFile(os.Getenv("IP_FORWARDING_PATH"), []byte("1\n"), 0644); err !=nil{
		return fmt.Errorf("Failed enable ip-forwarding: %v",err)
	}
	cmd := exec.Command("iptables", "-t", "nat", "-A", "POSTROUTING", "-s", internalNetwork, "!", "-o", "br0", "-j", "MASQUERADE")
    if err := cmd.Run(); err != nil{
		return fmt.Errorf("Failed enable masquerade with host: %v",err)
	}
	return nil	
}


func (net *networkManager) ShowVeth() []string{
	return net.createdVeths
}




func SetupContainerDNS(rootfsPath string) error {
	resolvConfPath := filepath.Join(rootfsPath, "etc", "resolv.conf")
	etcDir := filepath.Dir(resolvConfPath)
	if err := os.MkdirAll(etcDir, 0755); err != nil {
		return fmt.Errorf("failed to create /etc directory inside container rootfs: %w", err)
	}
	dnsConfig := []byte("nameserver 8.8.8.8\nnameserver 1.1.1.1\n")
	if err := os.WriteFile(resolvConfPath, dnsConfig, 0644); err != nil {
		return fmt.Errorf("failed to write /etc/resolv.conf inside container rootfs: %w", err)
	}
	return nil
}