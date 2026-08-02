package limits

import (
	"boyler/internal/daemon/core"
	"boyler/pkg/logger"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/containerd/cgroups/v3/cgroup2"
)


type ResourcesContainerManager interface {
	GroupOperator
	Apply(ctx context.Context, pid uint64) error
    Delete(ctx context.Context, pid uint64) error
	Update(ctx context.Context, res *core.Restriction) error
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
	rootSubtreeControlPath := filepath.Join(systemPath, "cgroup.subtree_control")
	if err := os.WriteFile(rootSubtreeControlPath, []byte("+cpu +memory"), 0644); err != nil {
		return nil, fmt.Errorf("failed to enable root subtree_control controllers: %w", err)
	}

	parentPath := filepath.Join(systemPath, cgroupPath)
	if err := os.MkdirAll(parentPath, 0755); err != nil {
		return nil, fmt.Errorf("failed to create parent cgroup directory %s: %w", parentPath, err)
	}

	subtreeControlPath := filepath.Join(parentPath, "cgroup.subtree_control")
	if err := os.WriteFile(subtreeControlPath, []byte("+cpu +memory"), 0644); err != nil {
		return nil, fmt.Errorf("failed to enable subtree_control controllers: %w", err)
	}

	manager := &resourcesContainerManager{
		cgroupPath:  cgroupPath,
		containerID: containerID,
		systemPath:  systemPath,
	}

	cgroupsManager, err := manager.createCgroupManager(res, containerID)
	if err != nil {
		return nil, err
	}
	operator := NewGroupOperator(cgroupsManager)
	manager.lowLevelManager = cgroupsManager
	manager.GroupOperator = operator
	return manager, nil
}

// SYSTEM_PATH="/sys/fs/cgroup"
// CGROUP_PATH="boyler_restriction"
func (c *resourcesContainerManager) createCgroupManager(res *core.Restriction, containerID string) (*cgroup2.Manager, error) {
	resources := mapResources(res)
	groupPath := "/" + filepath.Join(c.cgroupPath, containerID)
	manager, err := cgroup2.NewManager(c.systemPath, groupPath, resources)
	if err != nil {
		return nil, fmt.Errorf("failed to create cgroup manager for %s at %s: %w", groupPath, c.systemPath, err)
	}
	return manager, nil
}

func (c *resourcesContainerManager) Apply(ctx context.Context, pid uint64) error {
	log := logger.FromContext(ctx)
	log.Debug("start appliying parametrs to pid","pid",pid)
	err := c.lowLevelManager.AddProc(pid)
	if err != nil{
		return fmt.Errorf("Failed to add %d to cgropu: %v",pid, err)
	}
	log.Info("parametrs applied")
	return nil
}


func (c *resourcesContainerManager) Delete(ctx context.Context, pid uint64) error {
	log := logger.FromContext(ctx)
	log.Debug("start delete parametrs to pid","pid",pid)
	if err := c.lowLevelManager.Kill(); err != nil{
		log.Warn("failed to kill processes in cgroup", "error", err)
	}
	log.Debug("Kernel 2-second sleep, wait while cgroup-manager kill proc\n")
	time.Sleep(2*time.Second)
	if err := c.lowLevelManager.Delete(); err != nil {
		log.Error("failed to cgroup: %v","err", err)
		return fmt.Errorf("Failed to kill cgroup")
	}
	log.Info("cgroup successfully deleted")
	return nil
}


func (c *resourcesContainerManager) Update(ctx context.Context, res *core.Restriction) error{
	log := logger.FromContext(ctx)
	log.Debug("start cgroups update")
	resources := mapResources(res)
	err := c.lowLevelManager.Update(resources)
	if err != nil{
		return fmt.Errorf("Failed to update cgroups data: %v",err)
	}
	log.Info("changind applied")
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