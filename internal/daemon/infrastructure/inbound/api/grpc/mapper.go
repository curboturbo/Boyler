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

	if req.Name != "" {
		opts = append(opts, application.WithName(req.Name))
	}

	if len(req.Args) > 0 {
		opts = append(opts, application.WithArgs(req.Args))
	}

	if req.Resources != nil {
		limits := core.Restriction{}

		if req.Resources.Memory != nil && req.Resources.Memory.Exist {
			max := req.Resources.Memory.Max
			if max > 0 {
				limits.Memory.Max = &max
			}
		}
		if cpu := req.Resources.Cpu; cpu != nil {
			if cpu.Weight > 0 {
				if cpu.Weight > 10000 {
					weight := uint64(10000)
					limits.CPU.Weight = &weight
				} else {
					weight := cpu.Weight
					limits.CPU.Weight = &weight
				}
			}
			if cpu.Quota > 0 {
				quota := cpu.Quota
				limits.CPU.Quota = &quota

				period := cpu.Period
				if period == 0 {
					period = 100000
				}
				limits.CPU.Period = &period
			}
			if cpu.Cpus != "" {
				limits.CPU.Cpus = cpu.Cpus
			}
			if cpu.Mems != "" {
				limits.CPU.Mems = cpu.Mems
			}
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

func MapStopRequestToCommand(req *gen.StopRequest) application.StopContainerCommand {
	return application.StopContainerCommand{
		ContainerContext: application.ContainerContext{ID: req.ContainerId},
	}
}

func MapRemoveRequestToCommand(req *gen.RemoveRequest) application.RemoveContainerCommand {
	contCtx := application.ContainerContext{ID: req.ContainerId}
	return application.RemoveContainerCommand{
		ContainerContext: contCtx,
	}
}

func MapInsRequestToCommand(req *gen.InspectRequest) application.InspectContainerCommand {
	contCtx := application.ContainerContext{ID: req.ContainerId}
	return application.InspectContainerCommand{
		ContainerContext: contCtx,
	}
}