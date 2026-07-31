package cmd

import (
	"context"
	"fmt"
	"os"
	"text/tabwriter"

	pb "boyler/internal/daemon/infrastructure/inbound/api/grpc/gen"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(psCmd)
}

var psCmd = &cobra.Command{
	Use:   "ps",
	Short: "Show all containers",
	Run: func(cmd *cobra.Command, args []string) {

		client, conn, err := NewGrpcDaemonClient()
		if err != nil {
			fmt.Printf("Error connecting to daemon: %v\n", err)
			return
		}
		defer conn.Close()

		resp, err := client.ContainersList(
			context.Background(),
			&pb.PsRequest{},
		)
		if err != nil {
			fmt.Printf("Failed to fetch containers list: %v\n", err)
			return
		}
		printContainers(resp)
	},
}

func printContainers(resp *pb.PsResponse) {
	w := tabwriter.NewWriter(os.Stdout,0,0,3,' ',0,)
	fmt.Fprintln(
		w,
		"CONTAINER ID\tIMAGE\tCOMMAND\tCREATED\tSTATUS\tNAMES",
	)
	
	for _, container := range resp.GetContainers() {
		fmt.Fprintf(
			w,
			"%s\t%s\t%s\t%s\t%s\t%s\n",
			container.GetContainerId(),
			container.GetImage(),
			container.GetCommand(),
			container.GetCreated(),
			container.GetStatus(),
			container.GetName(),
		)
	}
	w.Flush()
}