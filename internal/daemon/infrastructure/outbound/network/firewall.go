package network

import (
	"boyler/pkg/workflow"
	"context"
	"fmt"
	"os/exec"
	"time"
)


type FirewallManager interface {
	Isolate(ipAddress string, bridgeName string) error
	CancelIsolation(ipAddress string, bridgeName string) error
	Forward(hostPort string, containerPort string, ipAddres string, bridgeName string) error
}

type firewallManager struct{
}

func NewFirewallManager() FirewallManager{
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
	wf := workflow.NewWorkflow()
	cmdContext, cancel := context.WithTimeout(context.Background(),time.Second*15)
	defer cancel()
	wf.Add(workflow.NewSagaStep(
		func() error {
			return routeLocalnetCmd(cmdContext,"1",bridgeName).Run()
		},
		func() error {
			return routeLocalnetCmd(cmdContext,"0",bridgeName).Run()
		},
	))
	wf.Add(workflow.NewSagaStep(
        func() error {
            return dnatPreroutingCmd(cmdContext, "-A", hostPort, ipAddres, containerPort).Run()
        	},
        func() error {
            return dnatPreroutingCmd(cmdContext,"-D", hostPort, ipAddres, containerPort).Run()
        	},
    	),
	)
	wf.Add(workflow.NewSagaStep(
		func() error {
			return allowForwardCmd(cmdContext,"-A",ipAddres,containerPort).Run()
		},
		func () error {
			return allowForwardCmd(cmdContext, "-D", ipAddres, containerPort).Run()
		},
	))
	wf.Add(workflow.NewSagaStep(
		func() error {
			return  localhostRedirectCmd(cmdContext, "-A", hostPort,ipAddres,containerPort).Run()
		},
		func () error {
			return localhostRedirectCmd(cmdContext, "-D",hostPort,ipAddres,containerPort).Run()
		},
	))
	wf.Add(workflow.NewSagaStep(
		func() error{
			return masqueradeCmd(cmdContext, "-A",bridgeName).Run()
		},
		func() error{
			return masqueradeCmd(cmdContext, "-D", bridgeName).Run()
		},
	))
	return wf.Execute()
}


func dnatPreroutingCmd(ctx context.Context, action, hostPort, ipAddress, containerPort string) *exec.Cmd {
	return exec.CommandContext(ctx, "iptables", "-t", "nat", action, "PREROUTING",
		"-p", "tcp", "--dport", hostPort,
		"-j", "DNAT", "--to-destination", fmt.Sprintf("%v:%v", ipAddress, containerPort))
}

func allowForwardCmd(ctx context.Context, action, ipAddress, containerPort string) *exec.Cmd {
	return exec.CommandContext(ctx, "iptables", action, "FORWARD",
		"-p", "tcp", "-d", ipAddress, "--dport", containerPort,
		"-j", "ACCEPT")
}

func localhostRedirectCmd(ctx context.Context, action, hostPort, ipAddress, containerPort string) *exec.Cmd {
	return exec.CommandContext(ctx, "iptables", "-t", "nat", action, "OUTPUT",
		"-p", "tcp", "-o", "lo", "--dport", hostPort,
		"-j", "DNAT", "--to-destination", ipAddress+":"+containerPort)
}

func masqueradeCmd(ctx context.Context,action, bridgeName string) *exec.Cmd {
	return exec.CommandContext(ctx, "iptables", "-t", "nat", action, "POSTROUTING",
		"-o", bridgeName, "-s", "127.0.0.1",
		"-j", "MASQUERADE")
}

func routeLocalnetCmd(ctx context.Context,value, bridgeName string) *exec.Cmd {
	return exec.CommandContext(ctx, "sysctl", "-w",
		fmt.Sprintf("net.ipv4.conf.%v.route_localnet=%v", bridgeName, value))
}
