package core

import "time"

type Container struct {
	ID        string
	PID       int
	ImageID   string
	CreatedAt time.Time
	StartedAt time.Time
	Config    ContainerConfig
}

type ContainerConfig struct {
	Hostname  string
	Env       []string
	Args      []string
	Resources Restriction
}

type Restriction struct {
	Memory MemoryRestriction
	CPU    CPURestriction
}

type MemoryRestriction struct {
	Max *int64
}

type CPURestriction struct {
	Weight *uint64
	Quota  *int64
	Period *uint64
	Cpus   string
	Mems   string
}
