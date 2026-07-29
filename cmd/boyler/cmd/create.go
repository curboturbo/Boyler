package cmd

import (
    "context"
    "fmt"
    "strings"
    "time"

    pb "boyler/internal/daemon/infrastructure/inbound/api/grpc/gen"
    utils "boyler/cmd/boyler/cmd/utils"
    "github.com/spf13/cobra"
)


var (
    containerName string
    envVars       []string
	ports         []string
    memoryLimit   string
    cpuShares     uint64
    cpuQuota      int64
    cpuPeriod     uint64
    cpusetCpus    string
    cpusetMems    string
)

func init() {
    rootCmd.AddCommand(createCmd)

    createCmd.Flags().StringVarP(&containerName, "name", "n", "", "Assign a name to the container")
    createCmd.Flags().StringSliceVarP(&envVars, "env", "e", []string{}, "Set environment variables")
	createCmd.Flags().StringSliceVarP(&ports, "publish", "p", []string{}, "Publish a container's port(s) to the host")

    createCmd.Flags().StringVarP(&memoryLimit, "memory", "m", "", "Memory limit (e.g., 512m, 1g)")
    createCmd.Flags().Uint64VarP(&cpuShares, "cpu-shares", "c", 0, "CPU shares (relative weight)")
    createCmd.Flags().Int64Var(&cpuQuota, "cpu-quota", 0, "Limit the CPU CFS quota")
    createCmd.Flags().Uint64Var(&cpuPeriod, "cpu-period", 0, "Limit the CPU CFS period")
    createCmd.Flags().StringVar(&cpusetCpus, "cpuset-cpus", "", "CPUs in which to allow execution (e.g., 0-3, 0,1)")
    createCmd.Flags().StringVar(&cpusetMems, "cpuset-mems", "", "MEMs in which to allow execution (NUMA-nodes)")
}

var createCmd = &cobra.Command{
    Use:   "create [IMAGE] [COMMAND...]",
    Short: "Create a new container",
    Args:  cobra.MinimumNArgs(1),
    Run: func(cmd *cobra.Command, args []string) {
        imageIdentity := args[0]
        var containerArgs []string
        if len(args) > 1 {
            containerArgs = args[1:]
        }

        envMap := make(map[string]string)
        for _, env := range envVars {
            key, val, found := strings.Cut(env, "=")
            if found {
                envMap[key] = val
            }
        }

        var memMax int64 = 0
        var memExist bool = false
        client, conn, err := NewGrpcDaemonClient()
        if err != nil {
            fmt.Printf("Error connecting to daemon: %v\n", err)
            return
        }
        defer conn.Close()
        req := &pb.CreateRequest{
            Name:          containerName,
            ImageIdentity: imageIdentity,
            Args:          containerArgs,
            Env:           envMap,
            Resources: &pb.ResourceLimits{
                Memory: &pb.MemoryRestriction{
                    Max:   memMax,
                    Exist: memExist,
                },
                Cpu: &pb.CPURestriction{
                    Weight: cpuShares,
                    Quota:  cpuQuota,
                    Period: cpuPeriod,
                    Cpus:   cpusetCpus,
                    Mems:   cpusetMems,
                },
            },
        }

        ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
        defer cancel()

        res, err := client.CreateContainer(ctx, req)
        if err != nil {
            fmt.Printf("Failed to create container: %v\n", err)
            return
        }
        fmt.Printf("Container running\n")
        utils.PrintProtoJSON(res)
    },
}

//x86_64-pc-linux-gnu