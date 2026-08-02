package limits

import (
	"boyler/internal/daemon/core"
)


type Factory interface {
	New(cgroupPath, containerID string, res *core.Restriction, systemPath string) (ResourcesContainerManager, error)
}

type defaultFactory struct{}

func NewFactory() Factory {
	return &defaultFactory{}
}

func (defaultFactory) New(cgroupPath, containerID string, res *core.Restriction, systemPath string) (ResourcesContainerManager, error) {
	return NewResourcesManager(cgroupPath, containerID, res, systemPath)
}