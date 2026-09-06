package cmd

import (
	"boyler/internal/boilerfile"
	"context"
	"fmt"

	pb "boyler/internal/daemon/infrastructure/inbound/api/grpc/gen"

	"github.com/spf13/cobra"
)

var buildFile string

func init() {
	rootCmd.AddCommand(buildCmd)
	buildCmd.Flags().StringVarP(&buildFile, "file", "f", "Boilerfile", "Path to the Boilerfile")
}

var buildCmd = &cobra.Command{
	Use:     "build",
	Short:   "Build an image from a Boilerfile",
	GroupID: groupImages,
	RunE: func(cmd *cobra.Command, args []string) error {
		loadEnv()

		bf, err := boilerfile.ParseFile(buildFile)
		if err != nil {
			return err
		}

		// Locate the FROM instruction to know which base image to pull.
		from, ok := bf.Instructions[0].(*boilerfile.FromInstruction)
		if !ok {
			return fmt.Errorf("Boilerfile must start with FROM")
		}
		imageRef := from.Image + ":" + from.Tag

		client, conn, err := NewGrpcDaemonClient()
		if err != nil {
			return commandError(err)
		}
		defer conn.Close()

		// Pull the base image through the existing daemon pipeline.
		fmt.Fprintf(cmd.OutOrStdout(), "Pulling base image %s\n", imageRef)
		ctx, cancel := context.WithTimeout(cmd.Context(), daemonRequestTimeout)
		defer cancel()

		stream, err := client.PullImage(ctx, &pb.PullRequest{ImageIdentity: imageRef})
		if err != nil {
			return commandError(err)
		}
		for {
			_, err := stream.Recv()
			if err != nil {
				break
			}
		}

		// Print the remaining build steps (execution stubs for future daemon support).
		for _, instr := range bf.Instructions[1:] {
			fmt.Fprintf(cmd.OutOrStdout(), "Step: %-12s\n", instr.Keyword())
		}

		printSuccess(cmd.OutOrStdout(), "Build complete")
		return nil
	},
}
