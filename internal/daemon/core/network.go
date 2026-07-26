package core

type PortMapping struct {
	HostPort      int
	ContainerPort int
	Protocol      string
}

type Net struct {
	Ports        []PortMapping
	IPAddress    string
	BridgeName   string
	HostVethName string
}
