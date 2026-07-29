package application

import (
	"boyler/internal/daemon/core"
)

type CreateContainerCommand struct {
	ContainerName string
	ImageName string
	Hostname  string
	Env       []string
	Args      []string
	Limits    core.Restriction
}

type CreateContainerOption func(*CreateContainerCommand)

func WithImage(image string) CreateContainerOption {
	return func(c *CreateContainerCommand) {
		c.ImageName = image
	}
}

func WithHostname(hostname string) CreateContainerOption {
	return func(c *CreateContainerCommand) {
		c.Hostname = hostname
	}
}

func WithEnv(env []string) CreateContainerOption {
	return func(c *CreateContainerCommand) {
		c.Env = env
	}
}

func WithArgs(args []string) CreateContainerOption {
	return func(c *CreateContainerCommand) {
		c.Args = args
	}
}

func WithLimits(limits core.Restriction) CreateContainerOption {
	return func(c *CreateContainerCommand) {
		c.Limits = limits
	}
}

func WithName(name string) CreateContainerOption {
	return func(c *CreateContainerCommand) {
		c.ContainerName = name
	}
}

func NewCreateContainerCommand(opts ...CreateContainerOption) CreateContainerCommand {
	memoryLimit := int64(256 * 1024 * 1024)
	cpuWeight := uint64(100)

	cmd := CreateContainerCommand{
		ImageName: "alpine", // first reliaze support only alpine image in container
		Limits: core.Restriction{
			Memory: core.MemoryRestriction{
				Max: &memoryLimit,
			},
			CPU: core.CPURestriction{
				Weight: &cpuWeight,
			},
		},
	}

	for _, opt := range opts {
		opt(&cmd)
	}

	return cmd
}


type StartContainerCommand struct {
	ContainerContext
}

type StopContainerCommand struct {
	ContainerContext
}

type RemoveContainerCommand struct {
	ContainerContext
}

type RestartContainerCommand struct {
	ContainerContext
}


type AttachContainerCommand struct {
	ContainerContext
}

type PauseContainerCommand struct {
	ContainerContext
}

type UnpauseContainerCommand struct {
	ContainerContext
}

type InspectContainerCommand struct {
	ContainerContext
}

type OutputType string

const (
	OutputStdout OutputType = "STDOUT"
	OutputStderr OutputType = "STDERR"
	OutputExit   OutputType = "EXIT"
)

type ContainerOutput struct {
	Type     OutputType
	Data     []byte
	ExitCode int32
}