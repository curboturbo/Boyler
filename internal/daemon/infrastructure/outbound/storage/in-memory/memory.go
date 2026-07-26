package storage

import (
	"context"
	"fmt"
	"sync"

	core "boyler/internal/daemon/core"
)

type ContainerRepository struct {
	mu   sync.RWMutex
	data map[string]core.Container
}

func NewContainerRepository() *ContainerRepository {
	return &ContainerRepository{
		data: make(map[string]core.Container),
	}
}

func (r *ContainerRepository) Save(ctx context.Context, container core.Container) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.data[container.ID] = container
	return nil
}

func (r *ContainerRepository) Get(ctx context.Context, id string) (core.Container, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	c, ok := r.data[id]
	if !ok {
		return core.Container{}, fmt.Errorf("container %s not found", id)
	}
	return c, nil
}

func (r *ContainerRepository) List(ctx context.Context) ([]core.Container, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	result := make([]core.Container, 0, len(r.data))
	for _, c := range r.data {
		result = append(result, c)
	}
	return result, nil
}

func (r *ContainerRepository) Delete(ctx context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.data, id)
	return nil
}

func (r *ContainerRepository) Update(ctx context.Context, container core.Container) error {
	return r.Save(ctx, container)
}