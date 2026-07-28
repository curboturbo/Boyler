package cmd

import (
	service "boyler/internal/daemon/application/container_service"
	net "boyler/internal/daemon/application/network_service"
	core "boyler/internal/daemon/core"
	layer "boyler/internal/daemon/infrastructure/outbound/image"
	limits "boyler/internal/daemon/infrastructure/outbound/limits"
	net_manager "boyler/internal/daemon/infrastructure/outbound/network"
	overlay "boyler/internal/daemon/infrastructure/outbound/overlay"
	r "boyler/internal/runtime/myrunc"
	logger "boyler/pkg/logger"
	registry "boyler/internal/daemon/infrastructure/outbound/registry"
	storage "boyler/internal/daemon/infrastructure/outbound/storage/in-memory"
	"context"
	"fmt"

	//"fmt"
	"os"
	"path/filepath"

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
		wd, err := os.Getwd()
		if err != nil {
			return
		}

		projectRoot := wd // /home/tema/Boyler
		if filepath.Base(wd) == "bin" {
			projectRoot = filepath.Dir(wd)
		}
		imagesPath := filepath.Join(projectRoot, "lib", "images")      // ~/Boyler/lib/images
		containersPath := filepath.Join(projectRoot,"lib", "containers") // ~/Boyler/lib/containers
		runtimeBinPath := filepath.Join(projectRoot, "bin/myrunc")

		fmt.Print("\n")

		fmt.Print(projectRoot)
		a := 2
		b := 1
		if a<b{}else{

		im := layer.NewImageManager(imagesPath)
		ru := r.NewMyRunc(runtimeBinPath)
		fs := overlay.NewOverlayManager(imagesPath, containersPath)

		config := net_manager.Config{
			Eth0: os.Getenv("DEFAULT_ETH0"),
			Forward: os.Getenv("IP_FORWARDING_PATH"),

		}

		network_manager := net_manager.NewNetworkManager(config)
		network_service_config := net.NetworkServiceConfig{
			BridgeName: os.Getenv("BRIDGE_NAME"),
			BridgeIP: os.Getenv("BRIDGE_IP"),
			InternalNetwork: os.Getenv("CONTAINER_LOCAL_NETWORK"),
		}

		network, _ := net.NewNetworkService(network_manager,network_service_config)
		grpc_daemon := service.NewContainerService(service.Deps{
			Runtime: ru,
			FS:      fs,
			Images:  im,
			Network: network,
			Reg:     registry.NewRepo(),
			Store:   storage.NewContainerRepository(),
			Logger:  logger.InitLogger(false),
			CgroupFactory: limits.NewFactory(),
			Conf: service.ServiceConfig{
				UnpackDir: os.Getenv("UNPACK_DIR"),
				ContainerDir: os.Getenv("CONTAINER_DIR"),
				CgroupPath: os.Getenv("CGROUP_PATH"),
				SystemPath: os.Getenv("SYSTEM_PATH"),
			},
		})

		containerRequest := service.CreateContainerCommand{
    		ContainerName: "alpine-test",
    		ImageName:     "alpine",
    		Hostname:      "alpine-box",
    		Env:           []string{"PATH=/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin"},
    		Args:          []string{"/bin/sh"},
    		Limits: core.Restriction{
        		Memory: core.MemoryRestriction{
            		Max: nil, // без жесткого лимита памяти (или укажите pointer на байты, например: int64(512 * 1024 * 1024))
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
		my_honest_testing_id :="8e6ff240-a1a5-4642-a58e-1a53b3222ee0"
		grpc_daemon.CreateAndStart(context.Background(), containerRequest)
		fmt.Printf("=======================-=-=-=-=-=-=-=-=-==-=-=-=-=-===================================\n")
		var mock_flag string
		fmt.Scan(&mock_flag)
		fmt.Printf("ТЕСТИРУЕМ останвоку НАШЕГО КОНТЙЕНЕРА\n")
		asn, err := grpc_daemon.Stop(context.Background(), service.StopContainerCommand{
			ContainerContext: service.ContainerContext{ID: my_honest_testing_id},
		})
		if err != nil{
			fmt.Printf("НЕ СМОГ УДАЛИТЬ: %v\n", err)
		}
		fmt.Printf("РЕСПОНС ОТ УДАЛЕНИЯ: %v",asn.ID)
		fmt.Printf("===========!!!!!!==@===@=======-=-=-=-=-=-=-=-=-==-=-=-=-=-=====@=========@===!!!!!!==================\n")

		fmt.Printf("===========!!!!!!==@===@=======-=-=-=-=-=-=-=-=-==-=-=-=-=-=====@=========@===!!!!!!==================\n")
		fmt.Printf("Запускаю тот же конйтенер заново\n")
		ans, err := grpc_daemon.Start(context.Background(), service.StartContainerCommand{
			ContainerContext: service.ContainerContext{ID: my_honest_testing_id},
		})
		if err != nil{
			fmt.Printf("НЕ СМОГ ЗАНОВО ЗАПУСТИТЬ ПРОЦЕСС: %v", err)
		}
		fmt.Print(ans.ID)
		}
	},

}