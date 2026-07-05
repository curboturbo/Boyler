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
	Resources Restriction `json:"cgroups"`
}


type Restriction struct {
    Memory MemoryRestriction `json:"memory"`
    CPU    CPURestriction    `json:"cpu"`
}

type MemoryRestriction struct {
    Max *int64 `json:"max"`
}

type CPURestriction struct {
    Weight *uint64 `json:"weight,omitempty"`
    Quota  *int64  `json:"quota,omitempty"`
    Period *uint64 `json:"period,omitempty"`
    Cpus string `json:"cpus,omitempty"`
    Mems string `json:"mems,omitempty"`
}