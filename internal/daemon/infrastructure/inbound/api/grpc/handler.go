package grpc

import (
	"boyler/internal/daemon/application/container_service"
	pb "boyler/internal/daemon/infrastructure/inbound/api/grpc/gen"
	"context"
	grpc "google.golang.org/grpc"
)

type DaemonHandler struct {
	containerService application.ContainerService
	pb.UnimplementedContainerServiceServer
}

func NewDaemonHandler(containerService application.ContainerService) *DaemonHandler {
	return &DaemonHandler{containerService: containerService}
}

func (d *DaemonHandler) CreateContainer(ctx context.Context, req *pb.CreateRequest) (*pb.CreateResponse, error) {
	command := MapCreateRequestToCommand(req)
	serviceResponse, err := d.containerService.CreateAndStart(ctx, command)
	if err != nil {
		return &pb.CreateResponse{}, err
	}
	return MapCreateResponceToProto(serviceResponse), nil
}

func (d *DaemonHandler) StartContainer(ctx context.Context, req *pb.StartRequest) (*pb.StartResponse, error) {
	command := MapStartRequestToCommand(req)
	serviceResponse, err := d.containerService.Start(ctx, command)
	if err != nil {
		return &pb.StartResponse{}, err
	}
	return MapStartResponceToProto(serviceResponse), nil
}

func (d *DaemonHandler) StopContainer(ctx context.Context, req *pb.StopRequest) (*pb.StopResponse, error) {
	command := MapStopRequestToCommand(req)
	serviceResponse, err := d.containerService.Stop(ctx, command)
	if err != nil {
		return &pb.StopResponse{}, err
	}
	return MapStopResponseToProto(serviceResponse), nil
}

func (d *DaemonHandler) RemoveContainer(ctx context.Context, req *pb.RemoveRequest) (*pb.RemoveResponse, error) {
	command := MapRemoveRequestToCommand(req)
	serviveResponse, err := d.containerService.Remove(ctx, command)
	if err != nil{
		return &pb.RemoveResponse{}, err
	}
	return MapRemoveResponseToProto(serviveResponse), nil
}


func (d *DaemonHandler) AttachContainer(req grpc.BidiStreamingServer[pb.AttachRequest, pb.AttachResponse]) error {
	return nil
}


func (d *DaemonHandler) InspectContainer(ctx context.Context, req *pb.InspectRequest) (*pb.InspectResponse, error) {
	command := MapInsRequestToCommand(req)
	serviceResponse, err := d.containerService.Inspect(ctx, command)
	if err != nil {
		return &pb.InspectResponse{}, err
	}
	return MapInspectResponseToProto(serviceResponse), nil
}