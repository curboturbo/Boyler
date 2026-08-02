package cmd

import (
	"context"
	"fmt"
	"io"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"

	"boyler/cmd/boyler/cmd/ui"
	pb "boyler/internal/daemon/infrastructure/inbound/api/grpc/gen"
)

func init() {
	rootCmd.AddCommand(pull)
}

var pull = &cobra.Command{
	Use:   "pull [IMAGE]",
	Short: "Pull image from docker registry",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		image := args[0]

		client, conn, err := NewGrpcDaemonPullingClient()
		if err != nil {
			return fmt.Errorf("connecting to daemon: %w", err)
		}
		defer conn.Close()

		ctx := context.Background()
		stream, err := client.PullImage(ctx, &pb.PullImageRequest{
			ImageIdentity: image,
		})
		if err != nil {
			return fmt.Errorf("creating grpc stream: %w", err)
		}

		events := make(chan tea.Msg, 100)
		go grpcReader(stream, image, events)

		p := tea.NewProgram(ui.New(events))
		_, err = p.Run()
		return err
	},
}

func grpcReader(stream grpc.ServerStreamingClient[pb.PullImageEvent],image string, events chan<- tea.Msg) {
	defer close(events)
	for {
		resp, err := stream.Recv()
		if err == io.EOF {
			events <- ui.DoneMsg{Image: image}
			return
		}
		if err != nil {
			events <- ui.DoneMsg{Image: image}
			return
		}

		events <- ui.ProgressMsg{
			ID:       resp.Layid,
			Status:   resp.Status,
			Progress: ratio(resp.Progress, resp.Total),
		}
	}
}

func ratio(current, total int64) float64 {
	if total <= 0 {
		return 0
	}
	if current >= total {
		return 1
	}
	return float64(current) / float64(total)
}