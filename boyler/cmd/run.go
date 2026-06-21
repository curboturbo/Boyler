package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(versionCmd)
}

var runCmd = &cobra.Command{
	Use:   "run [IMAGE]",
	Short: "Run container",
	Long: "Run container from loaded images or pull from DockerHub",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Boyler v0.1.0")
	},
}