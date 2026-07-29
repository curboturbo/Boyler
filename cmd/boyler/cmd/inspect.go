package cmd

import (
	"context"
	"fmt"
	"time"

	utils "boyler/cmd/boyler/cmd/utils"
	pb "boyler/internal/daemon/infrastructure/inbound/api/grpc/gen"

	"github.com/spf13/cobra"
)



func init() {
    rootCmd.AddCommand(inspectCmd)
}

var inspectCmd = &cobra.Command{
    Use:   "inspect [CONTAINER_ID]",
    Short: "Inspect container",
    Args:  cobra.MinimumNArgs(1),
    Run: func(cmd *cobra.Command, args []string) {
        id := args[0]
        client, conn, err := NewGrpcDaemonClient()
        if err != nil {
            fmt.Printf("Error connecting to daemon: %v\n", err)
            return
        }
        defer conn.Close()
        req := &pb.InspectRequest{ContainerId: id}
        ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
        defer cancel()

        res, err := client.InspectContainer(ctx, req)
        if err != nil {
            fmt.Printf("Failed to stop container: %v\n", err)
            return
        }
        fmt.Printf("Container metadata\n")
        utils.PrintProtoJSON(res)
    },
}