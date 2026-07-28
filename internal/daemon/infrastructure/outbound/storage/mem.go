package storage


import (
	"context"

	core "boyler/internal/daemon/core"
)

type ContainerStorage interface {
	Save(ctx context.Context, container core.Container) error
	Get(ctx context.Context, id string) (*core.Container, error)
	List(ctx context.Context) ([]*core.Container, error)
	Delete(ctx context.Context, id string) error
	Update(ctx context.Context, container core.Container) error
}