package limits

import (
	"fmt"
	"github.com/containerd/cgroups/v3/cgroup2"
)

type Stat struct {
	MemoryUsage uint64
	CPUUsage    uint64
	PIDCount    uint64
}

type GroupOperator interface {
	GetStats() (*Stat, error)
	Freeze() error
	Unfreeze() error
}

type operator struct{
	lowLevelMgr *cgroup2.Manager
}

func NewGroupOperator(lowLevelMgr *cgroup2.Manager) GroupOperator{
	return &operator{lowLevelMgr: lowLevelMgr}
}

func (oper *operator) Freeze() error {
	err := oper.lowLevelMgr.Freeze()
	if err != nil{
		return fmt.Errorf("Failed to freeze control group: %v", err)
	}
	return nil
}


func (oper *operator) Unfreeze() error { 
	err := oper.lowLevelMgr.Thaw()
	if err != nil {
		return fmt.Errorf("Failed to unfreeze control group: %v", err)
	}
	return nil
}

func (oper *operator) GetStats() (*Stat, error) { 
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
