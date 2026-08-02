package cmd

import (
    "context"
    "fmt"
    "time"

    pb "boyler/internal/daemon/infrastructure/inbound/api/grpc/gen"
    utils "boyler/cmd/boyler/cmd/utils"
    "github.com/spf13/cobra"
)



func init() {
    rootCmd.AddCommand(startCmd)
}

var startCmd = &cobra.Command{
    Use:   "start [CONTAINER_ID]",
    Short: "Start container",
    Args:  cobra.MinimumNArgs(1),
    Run: func(cmd *cobra.Command, args []string) {
        loadEnv()
        id := args[0]
        client, conn, err := NewGrpcDaemonClient()
        if err != nil {
            fmt.Printf("Error connecting to daemon: %v\n", err)
            return
        }
        defer conn.Close()
        req := &pb.StartRequest{ContainerId: id}
        ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
        defer cancel()

        res, err := client.StartContainer(ctx, req)
        if err != nil {
            fmt.Printf("Failed to start container: %v\n", err)
            return
        }
        fmt.Printf("Container started\n")
        utils.PrintProtoJSON(res)
    },
}