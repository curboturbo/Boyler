package grpc

import (
	"boyler/internal/daemon/application/container_service"
	"boyler/internal/daemon/core"
	pb "boyler/internal/daemon/infrastructure/inbound/api/grpc/gen"
	"fmt"

	grpc "google.golang.org/grpc"
)

type grpcStream struct {
	stream grpc.BidiStreamingServer[pb.AttachRequest, pb.AttachResponse]
}

func (g *grpcStream) Send(containerEvent *application.AttachOutboundEvent) error {
	resp := MapAttachResponse(containerEvent)
	return g.stream.Send(resp)
}

func (g *grpcStream) Receive() (*application.AttachInboundEvent, error) {
	req, err := g.stream.Recv()
	if err != nil {
		return nil, fmt.Errorf("Failed to receive grpc stream: %v", err)
	}
	return MapAttachRequest(req), nil
}


type grpcProgressStream struct {
	stream pb.ImageService_PullImageServer
}

func (g *grpcProgressStream) Send(event *core.PullingEvent) error {
	grcpEvent := MapCoreToProtoEvent(event)
	return g.stream.Send(grcpEvent)
}

type Stream interface{
	Send(*core.PullingEvent) error
}