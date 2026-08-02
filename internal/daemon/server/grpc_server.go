package server

import (
	grpchandler "boyler/internal/daemon/infrastructure/inbound/api/grpc"
	pb "boyler/internal/daemon/infrastructure/inbound/api/grpc/gen"
	inter "boyler/internal/daemon/infrastructure/inbound/api/grpc/interceptor"
	"fmt"
	"log/slog"
	"net"
	"os"
	"path/filepath"
	"time"

	"google.golang.org/grpc"
)


type Server struct {
	grpcServer *grpc.Server
	socketPath string
}

func NewGrpcServer(socketPath string, daemonHandler *grpchandler.DaemonHandler) *Server {
	grpcServer := grpc.NewServer(grpc.UnaryInterceptor(inter.ContextInterceptor(15 * time.Second)))
	pb.RegisterContainerServiceServer(grpcServer, daemonHandler)
	pb.RegisterImageServiceServer(grpcServer, daemonHandler)
	return &Server{
		grpcServer: grpcServer,
		socketPath: socketPath,
	}
}

func (s *Server) Start() error {
	if err := os.MkdirAll(parentDir(s.socketPath), 0755); err != nil {
		return fmt.Errorf("Failed mkdir: %v", err)
	}
	if err := os.Remove(s.socketPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("Failed to remove old unix-socket: %v", err)
	}
	lis, err := net.Listen("unix", s.socketPath)
	if err != nil {
		return fmt.Errorf("Failed to listen unix-socket: %v", err)
	}
	return s.grpcServer.Serve(lis)
}


func (s *Server) Stop() {
	if s.grpcServer != nil {
		s.grpcServer.GracefulStop()
	}
	if err := os.Remove(s.socketPath); err != nil && !os.IsNotExist(err) {
		slog.Error("Failed to delete unix-socket during shutdown", slog.String("error", err.Error()))
	}
}

func parentDir(socketPath string) string {
    return filepath.Dir(socketPath)
}