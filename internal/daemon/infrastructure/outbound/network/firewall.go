package network

import (
	"fmt"
	"os/exec"
)


type FirewallManager interface {
	Isolate(ipAddress string, bridgeName string) error
	CancelIsolation(ipAddress string, bridgeName string) error
	Forward(hostPort string, containerPort string, ipAddres string, bridgeName string) error
}

type firewallManager struct{

}

func NewFirewallManager () FirewallManager{
	return &firewallManager{}
}


func (f *firewallManager) Isolate(ipAddress string, bridgeName string) error {
	cmd  := exec.Command("iptables","-I","FORWARD","-s",ipAddress,"!","-o",bridgeName,"-j","DROP")
	if err := cmd.Run(); err != nil{
		return fmt.Errorf("Failed isolate ip - %v : %v", ipAddress, err)
	}
	return nil
}


func (f *firewallManager) CancelIsolation(ipAddress string, bridgeName string) error {
	cmd := exec.Command("iptables","-D","FORWARD","-s",ipAddress,"!","-o",bridgeName,"-j","DROP")
	if err := cmd.Run(); err != nil{
		return fmt.Errorf("Failed isolate ip - %v : %v", ipAddress, err)
	}
	return nil
}


func (f *firewallManager) Forward(hostPort string, containerPort string, ipAddres string, bridgeName string) error {
	cmd := exec.Command("iptables","-t","nat","-A","PREROUTING","-p","tcp","--dport",hostPort,"-j","DNAT","--to-destination",fmt.Sprintf("%v:%v",ipAddres, containerPort))
	if err := cmd.Run(); err != nil{
		return fmt.Errorf("Failed to translate connection from port %v to %v:%v",hostPort,ipAddres,containerPort)
	}

	allowForwardCmd := exec.Command("iptables", "-A", "FORWARD", 
		"-p", "tcp", "-d", ipAddres, "--dport", containerPort, 
		"-j", "ACCEPT")
	if err := allowForwardCmd.Run(); err != nil {
		return fmt.Errorf("failed to setup FORWARD rule: %v", err)
	}

	allowLocalCmd := exec.Command("sysctl", "-w", fmt.Sprintf("net.ipv4.conf.%v.route_localnet=1", bridgeName))
	if err := allowLocalCmd.Run(); err != nil {
		return fmt.Errorf("Failed to set up transition localhost traffic: %v",err)
	}

	localHostCmd := exec.Command("iptables","-t","nat","-A","OUTPUT","-p","tcp","-o","lo","--dport",hostPort,"-j","DNAT","--to-destination",ipAddres+":"+containerPort)
	if err := localHostCmd.Run(); err != nil{
		return fmt.Errorf("Failed to redirect localhost traffic to container port: %v",err)
	}

	masqueradeCmd := exec.Command("iptables", "-t", "nat", "-A", "POSTROUTING", 
		"-o", bridgeName, "-s", "127.0.0.1", 
		"-j", "MASQUERADE")
	if err := masqueradeCmd.Run(); err != nil {
		return fmt.Errorf("failed to setup localhost masquerade: %v", err)
	}

	return nil
}




func dnatPreroutingCmd(action, hostPort, ipAddress, containerPort string) *exec.Cmd {
	return exec.Command("iptables", "-t", "nat", action, "PREROUTING",
		"-p", "tcp", "--dport", hostPort,
		"-j", "DNAT", "--to-destination", fmt.Sprintf("%v:%v", ipAddress, containerPort))
}

func allowForwardCmd(action, ipAddress, containerPort string) *exec.Cmd {
	return exec.Command("iptables", action, "FORWARD",
		"-p", "tcp", "-d", ipAddress, "--dport", containerPort,
		"-j", "ACCEPT")
}

func localhostRedirectCmd(action, hostPort, ipAddress, containerPort string) *exec.Cmd {
	return exec.Command("iptables", "-t", "nat", action, "OUTPUT",
		"-p", "tcp", "-o", "lo", "--dport", hostPort,
		"-j", "DNAT", "--to-destination", ipAddress+":"+containerPort)
}

func masqueradeCmd(action, bridgeName string) *exec.Cmd {
	return exec.Command("iptables", "-t", "nat", action, "POSTROUTING",
		"-o", bridgeName, "-s", "127.0.0.1",
		"-j", "MASQUERADE")
}

func routeLocalnetCmd(value, bridgeName string) *exec.Cmd {
	return exec.Command("sysctl", "-w",
		fmt.Sprintf("net.ipv4.conf.%v.route_localnet=%v", bridgeName, value))
}
