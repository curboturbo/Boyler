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

        // 1. Получаем путь к текущему исполняемому файлу (например, boyler)
        exePath, err := os.Executable()
        if err != nil {
            fmt.Printf("Failed to get executable path: %v\n", err)
            return
        }
        resPath, _ := filepath.EvalSymlinks(exePath)
        binDir := filepath.Dir(resPath)
        
        // 2. Корень проекта — это папка на уровень выше bin/ (т.е. Boyler/)
        projectRoot := filepath.Dir(binDir)

        // 3. Формируем кроссплатформенный путь к демону
        daemonPath := filepath.Join(binDir, "daemon_boyler_linux") // или учтите ОС, если нужно

        cmdCommand := exec.Command(daemonPath)
        
        // 4. Главное: устанавливаем рабочую директорию демона в корень проекта
        cmdCommand.Dir = projectRoot

        if err := cmdCommand.Start(); err != nil {
            fmt.Printf("Fail to connect boyler daemon: %v\n", err)
            return
        }
        fmt.Printf("Boyler daemon is running ...\n")
    },
}