package cmd

import (
	"fmt"
	daemon "boyler/internal/daemon/CRI"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(versionCmd)

}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Show current version of app",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Boyler v0.1.0")
	},
}


func init() {
	rootCmd.AddCommand(checkCmd)
	
}

var checkCmd = &cobra.Command{
	Use:   "check",
	Short: "Show current version of app",
	Run: func(cmd *cobra.Command, args []string) {
		daemon.Check()
	},
}