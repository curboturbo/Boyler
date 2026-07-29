/*
Copyright © 2026 Arrdin <arrdin32@gmail.com>
*/
package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

const boylerLogo = `
  ____              _             
 | __ )  ___  _   _| | ___ _ __   
 |  _ \ / _ \| | | | |/ _ \ '__|  
 | |_) | (_) | |_| | |  __/ |     
 |____/ \___/ \__, |_|\___|_|     
              |___/               
        _ 
      (o\<    [ Lightweight Container Runtime ]
      //\\
      V_/_
`

const boylerManual = `
--- BOYLER CLI MANUAL ---

Overview:
  Boyler is a lightweight container engine built from scratch in Go 
  using Linux namespaces.

Available Commands:
  create       Create a new container (e.g., boyler create alpine --name HELLO_WORLD)
  start        Start an existing container
  stop         Stop a running container
  remove (rm)  Remove a container
  inspect      Show detailed information about a container
  version      Print the current version of Boyler

Architecture & Runtime:
  - Uses Linux namespaces (PID, Network, Mount, UTS, IPC) for isolation.
  - Applies resource restrictions (Memory, CPU weight/quota/period) via cgroups.
  - Communicates via gRPC between CLI and daemon (boylerd).
  - Built with Clean Architecture (gRPC inbound -> Application use cases -> Core domain -> Outbound drivers / myrunc).

Use "boyler [command] --help" for more information about a specific command.
`


// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "boyler",
	Short: "Boyler is a lightweight custom container engine and runtime",
	Long: `Boyler is a lightweight container engine built from scratch in Go 
using Linux namespaces, cgroups, and OCI-compatible principles. 
It allows you to create, start, and manage isolated containers with ease.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf(boylerLogo)
		cmd.Help()
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.SetHelpTemplate(`{{.Long}}

Usage:{{if .Runnable}}
  {{.UseLine}}{{end}}{{if .HasAvailableSubCommands}}
  {{.CommandPath}} [command]{{end}}{{if .HasAvailableFlags}}

Flags:
{{.LocalFlags.FlagUsages | trimTrailingWhitespaces}}{{end}}{{if .HasAvailableSubCommands}}

Available Commands:{{range .Commands}}{{if (or .IsAvailableCommand (eq .Name "help"))}}
  {{rpad .Name .NamePadding }} {{.Short}}{{end}}{{end}}{{end}}
%s
`)
}
