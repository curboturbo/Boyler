package application

import (
	"time"
	"boyler/internal/daemon/core"

	genName "boyler/pkg/generator"
)


type CoreContainerOption func(*core.Container)

func MapApplicationToCore(req *CreateContainerCommand, opts ...CoreContainerOption) *core.Container{
	newContainer := &core.Container{
		ImageID: req.ImageName,
		Status: core.StatusRunning,
		Config: core.ContainerConfig{
			Hostname: req.Hostname,
			Env: req.Env,
			Args: req.Args,
			Resources: req.Limits,
		},
	}
	for _, opt := range opts {
		opt(newContainer)
	}
	return newContainer	
}

func WithPid(pid int64) CoreContainerOption {
	return func(c *core.Container){
		c.PID = int(pid)
	}
}

func WithId(id string) CoreContainerOption {
	return func(c *core.Container) {
		c.ID = id
	}
}

func WithTime(create, start time.Time) CoreContainerOption{
	return func(c *core.Container) {
		c.CreatedAt, c.StartedAt = create, start
	}
}

func WithCoreName(name string) CoreContainerOption {
	return func(c *core.Container) {
		c.Name = genName.NameOrCreate(name)
	}
}