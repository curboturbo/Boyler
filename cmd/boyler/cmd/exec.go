package cmd

import (
	pb "boyler/internal/daemon/infrastructure/inbound/api/grpc/gen"
	"context"
	"fmt"
	"io"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"regexp"

	"github.com/spf13/cobra"
	"google.golang.org/grpc"
)

var (
	interactive bool
	tty         bool
)

var execCmd = &cobra.Command{
	Use:   "exec <container>",
	Short: "Execute shell inside container and stream from daemon. To escape send SIGTERM signal (Ctrl+C)",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		loadEnv()
		containerID := args[0]
		client, conn, err := NewGrpcDaemonClient()
		if err != nil {
			fmt.Printf("Error connecting to daemon: %v\n", err)
			return
		} 
		defer conn.Close()

		streamCtx, cancel := context.WithCancel(context.Background())
		stream, err := client.AttachContainer(streamCtx)
		if err != nil{
			fmt.Printf("Error open stream with deamon: %v", err)
			cancel()
			return
		}

		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan,syscall.SIGTERM, os.Interrupt)
		go func() {
			<-sigChan
			fmt.Printf("Streaming end\n")
			cancel()
		}()

		var wg sync.WaitGroup
		wg.Add(2)
		go commandInput(stream, &wg, containerID, streamCtx)
		go receiveOutput(stream, &wg)
		wg.Wait()
	},
}

func init() {
	rootCmd.AddCommand(execCmd)
	execCmd.Flags().BoolVarP(&interactive, "interactive", "i", false, "Keep STDIN open")
	execCmd.Flags().BoolVarP(&tty, "tty", "t", false, "Allocate a pseudo-TTY")
}



func receiveOutput(stream grpc.BidiStreamingClient[pb.AttachRequest, pb.AttachResponse], wg *sync.WaitGroup) {
	defer wg.Done()
	for {
		resp, err := stream.Recv()
		if err == io.EOF {
			return
		}
		if err != nil {
			fmt.Printf("Streaming terminated")
			return
		}
		fmt.Printf("%s\n", cleanANSI(resp.GetStdout()))
	}
}



func commandInput(stream grpc.BidiStreamingClient[pb.AttachRequest, pb.AttachResponse],wg *sync.WaitGroup,id string,ctx context.Context){
	defer wg.Done()
	req := &pb.AttachRequest{
		Payload: &pb.AttachRequest_Init{
			Init: &pb.AttachInit{ContainerId: id},
		},
	}
	if err := stream.Send(req); err != nil {
		fmt.Printf("Init request sending error (attach_request): %v\n", err)
		return
	}
	readCh := make(chan mes)
	go readBuf(readCh)
	for {
		select {
		case <-ctx.Done():
			stream.CloseSend()
			return

		case msg := <-readCh:
			if !msg.ok {
				stream.CloseSend()
				return
			}
			if err := stream.Send(&pb.AttachRequest{
				Payload: &pb.AttachRequest_Stdin{
					Stdin: msg.buf,
				},
			}); err != nil {
				fmt.Printf("Streaming sending error: %v\n", err)
				return
			}
		}
	}
}

type mes struct {
	ok  bool
	buf []byte
}

func readBuf(ch chan mes) {
	for {
		buf := make([]byte, 1024)
		n, err := os.Stdin.Read(buf)
		if err != nil {
			ch <- mes{ok: false}
			return
		}
		ch <- mes{ok: true, buf: buf[:n]}
	}
}

var csiRegexp = regexp.MustCompile(`\x1b\[[0-9;?]*[ -/]*[@-~]`)

func cleanANSI(data []byte) []byte {
    return csiRegexp.ReplaceAll(data, []byte{})
}