package grpc

import (
	"boyler/internal/daemon/application/container_service"
	"boyler/internal/daemon/core"
	"boyler/internal/daemon/infrastructure/inbound/api/grpc/gen"
)

func MapCreateRequestToCommand(req *gen.CreateRequest) application.CreateContainerCommand {
	opts := []application.CreateContainerOption{}

	if req.ImageIdentity != "" {
		opts = append(opts, application.WithImage(req.ImageIdentity))
	}

	if req.Name != ""{
		opts = append(opts, application.WithName(req.Name))
	}


	if len(req.Args) > 0 {
		opts = append(opts, application.WithArgs(req.Args))
	}

	if req.Resources != nil {
		limits := core.Restriction{}
		if req.Resources.Memory != nil && req.Resources.Memory.Exist {
			limits.Memory.Max = &req.Resources.Memory.Max
		}
		if req.Resources.Cpu != nil {
			limits.CPU.Weight = &req.Resources.Cpu.Weight
			limits.CPU.Quota = &req.Resources.Cpu.Quota
			limits.CPU.Period = &req.Resources.Cpu.Period
			limits.CPU.Cpus = req.Resources.Cpu.Cpus
			limits.CPU.Mems = req.Resources.Cpu.Mems
		}
		opts = append(opts, application.WithLimits(limits))
	}

	return application.NewCreateContainerCommand(opts...)
}

func MapStartRequestToCommand(req *gen.StartRequest) application.StartContainerCommand {
	contCtx := application.ContainerContext{ID: req.ContainerId}
	return application.StartContainerCommand{
		ContainerContext: contCtx,
	}
}


func MapRemoveRequestToCommand(req *gen.RemoveRequest) application.RemoveContainerCommand {
	contCtx := application.ContainerContext{ID: req.ContainerId}
	return application.RemoveContainerCommand{
		ContainerContext: contCtx,
	}
}