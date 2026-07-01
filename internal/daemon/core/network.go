package core

type PortMapping struct {
	HostPort      int    `json:"host_port"`
	ContainerPort int    `json:"container_port"`
	Protocol      string `json:"protocol"`
}

type Net struct {
	Ports []PortMapping `json:"ports"`
	IPAddress    string `json:"ip_address"`
	BridgeName   string `json:"bridge_name"`
	HostVethName string `json:"host_veth_name"`
}