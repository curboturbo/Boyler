package cmd

import (
	daemon "boyler/internal/daemon/application/container_service"
	net "boyler/internal/daemon/application/network_service"
	core "boyler/internal/daemon/core"
	layer "boyler/internal/daemon/infrastructure/outbound/image"
	net_manager "boyler/internal/daemon/infrastructure/outbound/network"
	overlay "boyler/internal/daemon/infrastructure/outbound/overlay"
	registry "boyler/internal/daemon/infrastructure/outbound/registry"
	r "boyler/internal/runtime/myrunc"
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
		reg := registry.NewRepo()
		network, _ := net.NewNetworkService(network_manager,network_service_config)
		daemon := daemon.NewContainerService(
			ru,
			fs,
			im,
			network,
			reg,
		)
		cont := core.Container{ImageID: "alpine"}
		daemon.CreateAndStart(context.Background(),cont)
		}
	},
}