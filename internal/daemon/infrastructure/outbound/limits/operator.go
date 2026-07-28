package limits

import (
	"context"
	"fmt"
	"github.com/containerd/cgroups/v3/cgroup2"
	"boyler/pkg/logger"
)

type Stat struct {
	MemoryUsage uint64
	CPUUsage    uint64
	PIDCount    uint64
}

type GroupOperator interface {
	GetStats(ctx context.Context) (*Stat, error)
	Freeze(ctx context.Context) error
	Unfreeze(ctx context.Context) error
}

type operator struct{
	lowLevelMgr *cgroup2.Manager
}

func NewGroupOperator(lowLevelMgr *cgroup2.Manager) GroupOperator{
	return &operator{lowLevelMgr: lowLevelMgr}
}

func (oper *operator) Freeze(ctx context.Context) error {
	log := logger.FromContext(ctx)
	log.Debug("Start freeze cgroup")
	err := oper.lowLevelMgr.Freeze()
	if err != nil{
		return fmt.Errorf("Failed to freeze control group: %v", err)
	}
	log.Debug("Cgroup pause")
	return nil
}


func (oper *operator) Unfreeze(ctx context.Context) error { 
	log := logger.FromContext(ctx)
	log.Debug("Start unfreeze cgroup")
	err := oper.lowLevelMgr.Thaw()
	if err != nil {
		return fmt.Errorf("Failed to unfreeze control group: %v", err)
	}
	log.Debug("Cgroup is running again")
	return nil
}

func (oper *operator) GetStats(ctx context.Context) (*Stat, error) { 
	log := logger.FromContext(ctx)
	log.Debug("Request cgroup info")
	metric, err := oper.lowLevelMgr.Stat()
	if err != nil{
		return &Stat{}, fmt.Errorf("Failed to get metrics control group:%v", err)
	}
	statistic := &Stat{}
	if metric.Memory != nil{
		statistic.MemoryUsage = metric.Memory.Usage
	}
	if metric.CPU != nil {
		statistic.CPUUsage = metric.CPU.UsageUsec
	}
	if metric.Pids != nil {
		statistic.PIDCount = metric.Pids.Current
	}
	return statistic, nil
}
