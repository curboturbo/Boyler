package grpc

import (
	"boyler/internal/daemon/application/container_service"
	pb "boyler/internal/daemon/infrastructure/inbound/api/grpc/gen"
	grpc "google.golang.org/grpc"
	"context"

)
type DaemonHandler struct{
	containerService application.ContainerService
	pb.UnimplementedContainerServiceServer
}


func (d *DaemonHandler) CreateContainer(ctx context.Context, req *pb.CreateRequest) (*pb.CreateResponse, error) {
	command := MapCreateRequestToCommand(req)
	serviceResponse, err := d.containerService.CreateAndStart(ctx, command)
	if err != nil{ return &pb.CreateResponse{}, err }
	return MapCreateResponceToProto(serviceResponse), nil
}

func (d *DaemonHandler) StartContainer(ctx context.Context, req *pb.StartRequest) (*pb.StartResponse, error) {
	command := MapStartRequestToCommand(req)
	serviceResponse, err := d.containerService.Start(ctx, command)
	if err != nil{ return &pb.StartResponse{}, err}
	return MapStartResponceToProto(serviceResponse), nil
}

func (d *DaemonHandler) StopContainer(ctx context.Context, req *pb.StopRequest) (*pb.StopResponse, error) {
	// TODO: логика остановки контейнера
	return &pb.StopResponse{}, nil
}

func (d *DaemonHandler) RemoveContainer(ctx context.Context, req *pb.RemoveRequest) (*pb.RemoveResponse, error) {
	// TODO: логика удаления контейнера
	return &pb.RemoveResponse{}, nil
}

func (d *DaemonHandler) InspectContainer(ctx context.Context, req *pb.InspectRequest) (*pb.InspectResponse, error) {
	// TODO: логика получения информации о контейнере
	return &pb.InspectResponse{}, nil
}


func (d *DaemonHandler) AttachContainer(req grpc.BidiStreamingServer[pb.AttachRequest, pb.AttachResponse]) error {
	return nil
}