package grpc

import (
	"boyler/internal/daemon/application/container_service"
	"boyler/internal/daemon/core"
	"boyler/internal/daemon/infrastructure/inbound/api/grpc/gen"
	"time"
)

func MapCreateResponceToProto(resp *application.CreateContainerResponse) *gen.CreateResponse {
	return &gen.CreateResponse{
		Status:      resp.Status,
		ContainerId: resp.ID,
	}
}

func MapStartResponceToProto(resp *application.StartContainerResponse) *gen.StartResponse {
	return &gen.StartResponse{
		ContainerId: resp.ID,
		Pid:         int32(resp.PID),
	}
}

func MapStopResponseToProto(resp *application.StopContainerResponse) *gen.StopResponse {
	return &gen.StopResponse{ContainerId: resp.ID}
}


func MapRemoveResponseToProto(resp *application.RemoveContainerResponse) *gen.RemoveResponse {
	return &gen.RemoveResponse{ContainerId: resp.ID}
}

func MapInspectResponseToProto(resp *application.InspectContainerResponse) *gen.InspectResponse {
	return &gen.InspectResponse{
		ContainerId: resp.ContainerID,
		Pid: resp.Pid,
		ImageId: resp.ImageID,
		CreatedAt: resp.CreatedAt,
		StartedAt: resp.StartedAt,
		Env: resp.Env,
		Args: resp.Args,
		Status: resp.Status,
		Hostname: resp.Hostname,
		Resources: &gen.ResourceLimits{
			Memory: &gen.MemoryRestriction{
				Max: *resp.Resources.Memory.Max,
				Exist: true,
			},
			Cpu: &gen.CPURestriction{
				Weight: *resp.Resources.CPU.Weight,
				Quota: *resp.Resources.CPU.Quota,
				Period: *resp.Resources.CPU.Period,
				Cpus: resp.Resources.CPU.Cpus,
				Mems: resp.Resources.CPU.Mems,
			},
		},
	}
}

func MapAttachResponse(ev *application.AttachOutboundEvent) *gen.AttachResponse {
	return &gen.AttachResponse{
		Payload: &gen.AttachResponse_Stdout{Stdout: ev.Stdout},
	}
}

func MapPsResponseToProto(resp []*core.Container) []*gen.ContainerListItem {
	items := make([]*gen.ContainerListItem, 0, len(resp))
	for _, container := range resp {
		items = append(items, &gen.ContainerListItem{
			ContainerId: container.ID,
			Image:       container.ImageID,
			Command:     "/bin/sh",
			Created:     container.CreatedAt.Format(time.RFC3339),
			Status:      string(container.Status),
			Name:        container.Name,
		})
	}
	return items
}

func MapCoreToProtoEvent(event *core.PullingEvent) *gen.PullImageEvent{
	return &gen.PullImageEvent{
		Status: event.Status,
		Layid: event.LayId,
		Progress: event.Progress,
		Total: event.Total,
	}
}