package cmd

import (
	"context"
	"fmt"
	service "boyler/internal/daemon/application/container_service"
	core "boyler/internal/daemon/core"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(runCmd)
}

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Run container",
	Long:  "Run container from loaded images or pull from DockerHub",
	Run: func(cmd *cobra.Command, args []string) {
		factory := NewDaemonFactoryFromEnv()
		grpc_daemon, err := factory.NewDaemon()
		if err != nil {
			fmt.Printf("Не удалось инициализировать демон: %v\n", err)
			return
		}

		containerRequest := service.CreateContainerCommand{
			ContainerName: "alpine-test",
			ImageName:     "alpine",
			Hostname:      "alpine-box",
			Env:           []string{"PATH=/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin"},
			Args:          []string{"/bin/sh"},
			Limits: core.Restriction{
				Memory: core.MemoryRestriction{
					Max: nil,
				},
				CPU: core.CPURestriction{
					Weight: nil,
					Quota:  nil,
					Period: nil,
					Cpus:   "",
					Mems:   "",
				},
			},
		}

		my_honest_testing_id := "8e6ff240-a1a5-4642-a58e-1a53b3222ee0"
		
		_, err = grpc_daemon.CreateAndStart(context.Background(), containerRequest)
		if err != nil {
			fmt.Printf("Ошибка при создании и запуске контейнера: %v\n", err)
		}

		fmt.Printf("=======================-=-=-=-=-=-=-=-=-==-=-=-=-=-===================================\n")
		var mock_flag string
		fmt.Scan(&mock_flag)
		fmt.Printf("ТЕСТИРУЕМ Остановку НАШЕГО КОНТЙЕНЕРА\n")

		asn, err := grpc_daemon.Stop(context.Background(), service.StopContainerCommand{
			ContainerContext: service.ContainerContext{ID: my_honest_testing_id},
		})
		if err != nil {
			fmt.Printf("НЕ СМОГ УДАЛИТЬ: %v\n", err)
		} else {
			fmt.Printf("РЕСПОНС ОТ УДАЛЕНИЯ: %v\n", asn.ID)
		}

		fmt.Printf("===========!!!!!!==@===@=======-=-=-=-=-=-=-=-=-==-=-=-=-=-=====@=========@===!!!!!!==================\n")
		fmt.Printf("Запускаю тот же конйтенер заново\n")

		ans, err := grpc_daemon.Start(context.Background(), service.StartContainerCommand{
			ContainerContext: service.ContainerContext{ID: my_honest_testing_id},
		})
		if err != nil {
			fmt.Printf("НЕ СМОГ ЗАНОВО ЗАПУСТИТЬ ПРОЦЕСС: %v\n", err)
			return
		}
		if ans != nil {
			fmt.Print(ans.ID)
		}
	},
}