package application

import "context"

type ContainerService interface {
	CreateAndStart(ctx context.Context, cmd CreateContainerCommand) (*CreateContainerResponse, error)
	Start(ctx context.Context, cmd StartContainerCommand) (*StartContainerResponse, error)
	Stop(ctx context.Context, cmd StopContainerCommand) (*StopContainerResponse, error)
	Pause(ctx context.Context, cmd PauseContainerCommand) (*PauseContainerResponse, error)
	Unpause(ctx context.Context, cmd UnpauseContainerCommand) (*UnpauseContainerResponse, error)
	Remove(ctx context.Context, cmd RemoveContainerCommand) (*RemoveContainerResponse, error)
	Restart(ctx context.Context, cmd RestartContainerCommand) (*RestartContainerResponse, error)
	Inspect(ctx context.Context, cmd InspectContainerCommand) (*InspectContainerResponse, error)
	Attach(ctx context.Context, cmd AttachContainerCommand) (*AttachSession, error) // most harders part all grpc
}

type containerService struct {
	creator   *Creator
	stopper   *Stopper
	pauser    *Pauser
	unpauser  *Unpauser
	remover   *Remover
	restarter *Restarter
	attacher  *Attacher
	cursor    *Cursor
}

func NewContainerService(d Deps) ContainerService {
	return &containerService{
		creator:   NewCreator(d),
		stopper:   NewStopper(d),
		pauser:    NewPauser(d),
		unpauser:  NewUnpauser(d),
		remover:   NewRemover(d),
		restarter: NewRestarter(d),
		attacher:  NewAttacher(d),
		cursor: NewCursor(d),
	}
}

func (c *containerService) CreateAndStart(ctx context.Context, cmd CreateContainerCommand) (*CreateContainerResponse, error) {
	return c.creator.ExecuteCreate(ctx, cmd)
}

func (c *containerService) Start(ctx context.Context, cmd StartContainerCommand) (*StartContainerResponse, error) {
	return c.creator.ExecuteStart(ctx, cmd)
}

func (c *containerService) Stop(ctx context.Context, cmd StopContainerCommand) (*StopContainerResponse, error) {
	return c.stopper.Execute(ctx, cmd)
}

func (c *containerService) Pause(ctx context.Context, cmd PauseContainerCommand) (*PauseContainerResponse, error) {
	return c.pauser.Execute(ctx, cmd)
}

func (c *containerService) Unpause(ctx context.Context, cmd UnpauseContainerCommand) (*UnpauseContainerResponse, error) {
	return c.unpauser.Execute(ctx, cmd)
}

func (c *containerService) Remove(ctx context.Context, cmd RemoveContainerCommand) (*RemoveContainerResponse, error) {
	return c.remover.Execute(ctx, cmd)
}

func (c *containerService) Restart(ctx context.Context, cmd RestartContainerCommand) (*RestartContainerResponse, error) {
	return c.restarter.Execute(ctx, cmd)
}

func (c *containerService) Attach(ctx context.Context, cmd AttachContainerCommand) (*AttachSession, error) {
	return c.attacher.Execute(ctx, cmd)
}

func (c *containerService) Inspect(ctx context.Context, cmd InspectContainerCommand) (*InspectContainerResponse, error) {
	return c.cursor.Execute(ctx, cmd)
}