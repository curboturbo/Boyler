package cmd

import (
	"os"

	pb "boyler/internal/daemon/infrastructure/inbound/api/grpc/gen"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func NewGrpcDaemonClient() (pb.ContainerServiceClient, *grpc.ClientConn, error) {
	target := "unix://" + os.Getenv("UNIX_SOCKET")
	connection, err := grpc.NewClient(target, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, nil, err
	}
	client := pb.NewContainerServiceClient(connection)
	return client, connection, nil
}