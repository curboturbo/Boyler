package grpc

import (
	"boyler/internal/daemon/application/container_service"
	"boyler/internal/daemon/infrastructure/inbound/api/grpc/gen"
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
