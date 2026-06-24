package runtime

import (
	"context"
	"os"
)

type Runtime interface {
	Create(ctx context.Context, id string, bundlePath string) error

	Start(ctx context.Context, id string) error

	Kill(ctx context.Context, id string, signal os.Signal) error

	Delete(ctx context.Context, id string) error

	State(ctx context.Context, id string) (*State, error)
}

type Status string

const (
	StatusCreating Status = "creating"
	StatusCreated  Status = "created"
	StatusRunning  Status = "running"
	StatusStopped  Status = "stopped"
)

type State struct {
	ID          string `json:"id"`
	Status      Status `json:"status"`
	PID         int    `json:"pid,omitempty"`
	BundlePath  string `json:"bundle"`
	Annotations map[string]string `json:"annotations"`
}