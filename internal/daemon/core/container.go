package core

import "time"


type Container struct {
	ID string `json:"id"`
	PID int `json:"pid"`
	ImageID string `json:"imageid"`
	CreatedAt time.Time `json:"created_at"`
	StartedAt time.Time `json:"started_at"`
	Config ContainerConfig `json:"config"`
}


type ContainerConfig struct {
	Hostname string `json:"hostname"`
	Env []string  	`json:"env"`
	Args []string	`json:"args"`
	Resources Cgroups `json:"cgroups"`
}


type Cgroups struct {
	MemoryLimit int64	`json:"memory_limit"`
	CPULimit int64 `json:"cpu_limit"`
	CPUfsQ int64 	`json:"cpu_quota"`
}