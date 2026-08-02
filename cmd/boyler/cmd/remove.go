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
	rootCmd.AddCommand(removeCmd)
}

var removeCmd = &cobra.Command{
	Use:   "remove [CONTAINER_ID]",
	Short: "Remove a container",
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

		req := &pb.RemoveRequest{ContainerId: id}
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		res, err := client.RemoveContainer(ctx, req)
		if err != nil {
			fmt.Printf("Failed to remove container: %v\n", err)
			return
		}
		fmt.Printf("Container removed\n")
		utils.PrintProtoJSON(res)
	},
}
