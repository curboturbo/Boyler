package application

import "context"

type Attacher struct{}

func NewAttacher(d Deps) *Attacher {
	return &Attacher{}
}

func (a *Attacher) Execute(ctx context.Context, cmd AttachContainerCommand) (*AttachSession, error) {
	return nil, nil
}