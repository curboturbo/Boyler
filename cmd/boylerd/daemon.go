package main

import (
	"fmt"
	"os"

	grpchandler "boyler/internal/daemon/infrastructure/inbound/api/grpc"
	server "boyler/internal/daemon/server"
	logger "boyler/pkg/logger"
)

func main() {
	factory := NewDaemonFactoryFromEnv()
	containerService, err := factory.NewContainerService()
	if err != nil { panic(fmt.Sprintf("CANNOT CREATE DEAMON: %v", err)) }
	imageService, err := factory.NewImageService()
	if err != nil { panic(fmt.Sprintf("CANNOT CREATE DEAMON: %v", err)) }
	pprofConf:= server.ServerConfig{
		Addr: os.Getenv("HTTP_PPROF_SOCKET"),
		Log: logger.InitLogger(true),
	}
	server.StartPprofServer(pprofConf)
	grpcServer := server.NewGrpcServer(os.Getenv("UNIX_SOCKET"), grpchandler.NewDaemonHandler(containerService, imageService))
	if err := grpcServer.Start(); err != nil {
		fmt.Fprintf(os.Stderr, "daemon stopped: %v\n", err)
	}
}