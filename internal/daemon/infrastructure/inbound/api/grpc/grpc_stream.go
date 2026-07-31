package grpc

import(
	pb "boyler/internal/daemon/infrastructure/inbound/api/grpc/gen"
	grpc "google.golang.org/grpc"
	"boyler/internal/daemon/application/container_service"
	"fmt"
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