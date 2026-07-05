package limits

import (
	"boyler/internal/daemon/core"
	"fmt"
	"github.com/containerd/cgroups/v3/cgroup2"
)


// using delegation pattern, brings responsibility to GroupOperator
type ResourcesContainerManager interface {
	GroupOperator
	Apply(pid uint64) error
    Delete(pid uint64) error
	Update(res *core.Restriction) error
}

// unique struct for each container
type resourcesContainerManager struct{
	cgroupPath string
	systemPath string
	containerID string
	lowLevelManager *cgroup2.Manager
	GroupOperator
}

func NewResourcesManager(cgroupPath string, containerID string, res *core.Restriction, systemPath string) (ResourcesContainerManager, error) {
	manager := &resourcesContainerManager{cgroupPath: cgroupPath,
		containerID: containerID,
		systemPath: systemPath,
	}
	cgroupsManager, err := manager.createCgroupManager(res, containerID)
	if err != nil{
		return nil, err
	}
	operator := NewGroupOperator(cgroupsManager)
	manager.lowLevelManager = cgroupsManager
	manager.GroupOperator = operator
	return manager, nil
}


func (c *resourcesContainerManager) createCgroupManager(res *core.Restriction, containerID string) (*cgroup2.Manager, error) {
	resources :=  mapResources(res)
	manager, err := cgroup2.NewManager(c.cgroupPath,containerID, resources)
	if err != nil{
		return nil, fmt.Errorf("Failed to create cgroup for container %v: %v", containerID, err)
	}
	return manager, nil
}


func (c *resourcesContainerManager) Apply(pid uint64) error {
	err := c.lowLevelManager.AddProc(pid)
	if err != nil{
		return fmt.Errorf("Failed to add %d to cgropu: %v",pid, err)
	}
	return nil
}

// systemPath = "/sys/fs/cgroup"
func (c *resourcesContainerManager) Delete(pid uint64) error {
	rootMgr, err := cgroup2.NewManager(c.systemPath, "/", nil)
	if err != nil {
		return fmt.Errorf("Failed to connect to root cgroup: %w", err)
	}
	if err := rootMgr.AddProc(pid); err != nil{
		return fmt.Errorf("Failed to add pid to root cgroups: %v",err)
	}
	return nil
}


func (c *resourcesContainerManager) Update(res *core.Restriction) error{
	resources := mapResources(res)
	err := c.lowLevelManager.Update(resources)
	if err != nil{
		return fmt.Errorf("Failed to update cgroups data: %v",err)
	}
	return nil
}


func mapResources(res *core.Restriction) *cgroup2.Resources {
    return &cgroup2.Resources{
        CPU: &cgroup2.CPU{
            Max: cgroup2.NewCPUMax(
                res.CPU.Quota,
                res.CPU.Period,
            ),
            Weight: res.CPU.Weight,
            Cpus:   res.CPU.Cpus,
            Mems:   res.CPU.Mems,
        },
        Memory: &cgroup2.Memory{
            Max: res.Memory.Max,
        },
    }
}