package runtime

// General runtime interface for each run (myrunc, runc, crun)


import (
	"context"
	"os"
)

type Runtime interface {
	Create(ctx context.Context, id string, bundlePath string) (*State, error)

	Run(ctx context.Context, id string) (*State, error)

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

const OCI_VERSION = "1.02.2"

type State struct {
	ID 			string `json:"id"`
	PID         int    `json:"pid"`
	OciVerion   string `json:"ociVersion"`
	Status      Status `json:"status"`
	BundlePath  string `json:"bundle"`
}