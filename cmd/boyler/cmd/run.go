package cmd

import (
	daemon "boyler/internal/daemon/CRI"
	"fmt"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(runCmd)
}

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Run container",
	Long: "Run container from loaded images or pull from DockerHub",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("Начинаю запуск контейнера")
		daemon.Run()
	},
}