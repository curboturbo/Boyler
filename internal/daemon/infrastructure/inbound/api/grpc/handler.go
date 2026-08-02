package grpc

import (
	"context"

	grpc "google.golang.org/grpc"

	application "boyler/internal/daemon/application/container_service"
	imageservice "boyler/internal/daemon/application/image_service"
	pb "boyler/internal/daemon/infrastructure/inbound/api/grpc/gen"
)

type DaemonHandler struct {
	containerService application.ContainerService
	pb.UnimplementedContainerServiceServer
	imageService imageservice.ImageService
	pb.UnimplementedImageServiceServer
}

func NewDaemonHandler(containerService application.ContainerService, imageService imageservice.ImageService) *DaemonHandler {
	return &DaemonHandler{containerService: containerService, imageService: imageService}
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

func (d *DaemonHandler) InspectContainer(ctx context.Context, req *pb.InspectRequest) (*pb.InspectResponse, error) {
	command := MapInsRequestToCommand(req)
	serviceResponse, err := d.containerService.Inspect(ctx, command)
	if err != nil {
		return &pb.InspectResponse{}, err
	}
	return MapInspectResponseToProto(serviceResponse), nil
}

func (d *DaemonHandler) AttachContainer(req grpc.BidiStreamingServer[pb.AttachRequest, pb.AttachResponse]) error {
	return d.containerService.Attach(req.Context(), &grpcStream{stream:req})
}

func (d *DaemonHandler) ContainersList(ctx context.Context, req *pb.PsRequest) (*pb.PsResponse, error) {
	command := application.PsCommand{}
	serviceResponse, err := d.containerService.PsInspect(ctx, command)
	if err != nil{
		return &pb.PsResponse{}, err
	}
	containers := MapPsResponseToProto(serviceResponse)
	return &pb.PsResponse{
		Containers: containers,
	}, nil
}

func (d *DaemonHandler) PullImage(req *pb.PullImageRequest, stream pb.ImageService_PullImageServer) error {
	ctx := stream.Context()
	return d.imageService.Pull(ctx, req.GetImageIdentity(), &grpcProgressStream{stream: stream})
}

func (d *DaemonHandler) RemoveImage(ctx context.Context, req *pb.RemoveImageRequest) (*pb.RemoveImageResponse, error) {
	// TODO: реализовать заглушку для удаления образа
	return &pb.RemoveImageResponse{}, nil
}

func (d *DaemonHandler) ListImages(ctx context.Context, req *pb.ListImagesRequest) (*pb.ListImagesResponse, error) {
	// TODO: реализовать заглушку для получения списка образов
	return &pb.ListImagesResponse{}, nil
}