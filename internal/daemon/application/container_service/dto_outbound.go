package application

import (
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

type DeleteContainerResponse struct {
	ContainerContext
}

type StopContainerResponse struct {
	ContainerContext
}

type AttachSession struct {
    Stdin  io.WriteCloser
    Stdout io.Reader
    Stderr io.Reader
    Wait   func() error
}
