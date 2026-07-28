package application

import "context"

type Restarter struct{}

func NewRestarter(d Deps) *Restarter {
	return &Restarter{}
}

func (r *Restarter) Execute(ctx context.Context, cmd RestartContainerCommand) (*RestartContainerResponse, error) {
	return nil, nil
}