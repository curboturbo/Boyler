package cmd

import (
    "context"
    "fmt"
    "time"

    pb "boyler/internal/daemon/infrastructure/inbound/api/grpc/gen"
    utils "boyler/cmd/boyler/cmd/utils"
    "github.com/spf13/cobra"
)


var (
    containerId string
)

func init() {
    rootCmd.AddCommand(stopCmd)
}

var stopCmd = &cobra.Command{
    Use:   "stop [CONTAINER_ID]",
    Short: "Stop container",
    Args:  cobra.MinimumNArgs(1),
    Run: func(cmd *cobra.Command, args []string) {
        id := args[0]
        client, conn, err := NewGrpcDaemonClient()
        if err != nil {
            fmt.Printf("Error connecting to daemon: %v\n", err)
            return
        }
        defer conn.Close()
        req := &pb.StopRequest{ContainerId: id}
        ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
        defer cancel()

        res, err := client.StopContainer(ctx, req)
        if err != nil {
            fmt.Printf("Failed to stop container: %v\n", err)
            return
        }
        fmt.Printf("Container stopped\n")
        utils.PrintProtoJSON(res)
    },
}