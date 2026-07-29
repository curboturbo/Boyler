package cmd

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"time"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(initDaemonCmd)
}

var initDaemonCmd = &cobra.Command{
	Use:   "init",
	Short: "Show current version of app",
	Run: func(cmd *cobra.Command, args []string) {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
		defer cancel()
		cmdCommand := exec.CommandContext(ctx, os.Getenv("BIN_DAEMON"))
		cmdCommand.Env = os.Environ()
		if err := cmdCommand.Start(); err != nil{
			fmt.Printf("Fail to connect boyler daemon: %v",err)
			return
		}
		fmt.Printf("Boyler daemon is running ...\n")
	},
}
