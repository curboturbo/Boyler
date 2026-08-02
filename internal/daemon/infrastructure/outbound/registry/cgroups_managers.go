package registry

import (
	lim "boyler/internal/daemon/infrastructure/outbound/limits"
	"fmt"
	"sync"
)



type ResourcesRegistry interface {
	Get(containerID string) (lim.ResourcesContainerManager, error)
	Post(containerID string, manager lim.ResourcesContainerManager) error
	List() []lim.ResourcesContainerManager
	Delete(containerID string) error
	Drop() error
}


type cgroupsRepository struct {
	mu sync.RWMutex
	repository map[string]lim.ResourcesContainerManager
}


func NewRepo() ResourcesRegistry {
	repository := make(map[string]lim.ResourcesContainerManager,15)
	return &cgroupsRepository{repository: repository, mu: sync.RWMutex{}}
}


func (repo *cgroupsRepository) Get(containerID string) (lim.ResourcesContainerManager, error) {
	repo.mu.RLock()
	defer repo.mu.RUnlock()
	value, status := repo.repository[containerID]
	if status {
		return value, nil
	}else{
		return nil, fmt.Errorf("Failed not found %v manager in repo",containerID)
	}
}


func (repo *cgroupsRepository) Post(containerID string,manager lim.ResourcesContainerManager) error {
	repo.mu.Lock()
	defer repo.mu.Unlock()
	repo.repository[containerID] = manager
	return nil
}


func (repo *cgroupsRepository) List() []lim.ResourcesContainerManager{
	repo.mu.RLock()
	defer repo.mu.RUnlock()
	list := make([]lim.ResourcesContainerManager, len(repo.repository))
	for _, v := range repo.repository{
		list = append(list, v)
	}
	return list
}

func (repo *cgroupsRepository) Delete(containerID string) error {
	repo.mu.Lock()
	defer repo.mu.Unlock()
	_, ok := repo.repository[containerID]
	if ok {delete(repo.repository, containerID)}else{
		return fmt.Errorf("Failed to find key")
	}
	return nil
}

func (repo *cgroupsRepository) Drop() error {
	repo.mu.Lock()
	defer repo.mu.Unlock()
	repo.repository = nil
	repo.repository = make(map[string]lim.ResourcesContainerManager,15)
	return nil
}