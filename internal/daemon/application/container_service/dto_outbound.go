package application

import (
	"boyler/internal/daemon/core"
	"io"
)

type ContainerContext struct {
	ID 		string
}

type CreateContainerResponse struct {
	ContainerContext
	Status 	  string
}

type StartContainerResponse struct {
	PID int64
	ContainerContext
}

type RestartContainerResponse struct {
	ContainerContext
}

type RemoveContainerResponse struct {
	ContainerContext
}

type StopContainerResponse struct {
	ContainerContext
}

type PauseContainerResponse struct {
	ContainerContext
}

type UnpauseContainerResponse struct {
	ContainerContext
}

type InspectContainerResponse struct {
	ContainerID string
	Pid         int32
	ImageID     string
	CreatedAt   string
	StartedAt   string
	Status      string
	Hostname    string
	Env         []string
	Args        []string
	Resources   core.Restriction
}

type AttachSession struct {
    Stdin  io.WriteCloser
    Stdout io.Reader
    Stderr io.Reader
    Wait   func() error
}
