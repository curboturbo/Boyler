package main

import (
	"fmt"
	"os"
)

type execInfo struct {
	binaryPath string // file to binary example
	id string		  // container id
	bundlePath string // path to boyler/container/container_xxx
}


func main(){
	if len(os.Args) < 3 {
		fmt.Fprintln(os.Stderr, "Wait for command: create, start, state, delete, kill")
		os.Exit(1)
	}
	command := os.Args[1]
	containerID := os.Args[2]
	switch command{
		case "create":
			if len(os.Args) < 5 || os.Args[3]!= "--bundle"{
				fmt.Fprintln(os.Stderr,"Error, you have to send --bundle <path>")
				os.Exit(1)
			}
			bundlePath := os.Args[4]
			execCreateContainer(&execInfo{
				binaryPath: os.Args[0],
				id: containerID,
				bundlePath: bundlePath,
			})

		case "start":
			execStartContainer(&execInfo{
				id:containerID,
				binaryPath: "",
				bundlePath: os.Args[4],
			})

		case "state":
			// чтение state.json

		case "init":
			execInitContainer(&execInfo{
				id: containerID,
				binaryPath: "",
				bundlePath: os.Args[4],
			})

		case "delete":
			// удаление

		case "kill":
			// syscall на процесс

		default:
			fmt.Fprintf(os.Stderr, "Unknown command: %s\n",command)
			os.Exit(1)
	}
}


