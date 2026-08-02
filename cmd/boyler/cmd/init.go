package cmd

import (
	"fmt"
	"os/exec"
	"path/filepath"
	"os"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(initDaemonCmd)
}

var initDaemonCmd = &cobra.Command{
    Use:   "init",
    Short: "Show current version of app",
    Run: func(cmd *cobra.Command, args []string) {
        loadEnv()
        exePath, err := os.Executable()
        if err != nil {
            fmt.Printf("Failed to get executable path: %v\n", err)
            return
        }
        resPath, _ := filepath.EvalSymlinks(exePath)
        binDir := filepath.Dir(resPath)
        projectRoot := filepath.Dir(binDir)
        daemonPath := filepath.Join(binDir, "daemon_boyler_linux")
        cmdCommand := exec.Command(daemonPath)
        cmdCommand.Dir = projectRoot
        if err := cmdCommand.Start(); err != nil {
            fmt.Printf("Fail to connect boyler daemon: %v\n", err)
            return
        }
        fmt.Printf("Boyler daemon is running ...\n")
    },
}