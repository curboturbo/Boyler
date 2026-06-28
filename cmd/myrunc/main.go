package main

import (
	"fmt"
	"os"
)

// info for start process
type execInfo struct {
	binaryPath string // file to binary example
	id string		  // container id
	bundlePath string // path to boyler/container/container_xxx
	sigNum string     // signal to process (SIGTERM, SIGKILL and e.t.c)
}


func main(){
	if len(os.Args) < 3 {
		fmt.Fprintln(os.Stderr, "Wait for command: create, start, state, delete, kill")
		os.Exit(1)
	}
	command := os.Args[1]
	containerID := os.Args[2]
	var err error
	switch command{
		case "create":
			if len(os.Args) < 5 || os.Args[3]!= "--bundle"{
				fmt.Fprintln(os.Stderr,"Error, you have to send --bundle <path>")
				os.Exit(1)
			}
			bundlePath := os.Args[4]
			err = execCreateContainer(&execInfo{
				binaryPath: os.Args[0],
				id: containerID,
				bundlePath: bundlePath,
			})
			
		case "run":
			err = execRunContainer(&execInfo{
				id:containerID,
			})

		case "state":
			err = execCheckStateContainer(&execInfo{
				id:containerID,
			})

		case "init":
			err = execInitContainer(&execInfo{
				id: containerID,
				bundlePath: os.Args[4],
			})

		case "stop":
			err = execKillContainer(&execInfo{
				id:containerID,
				sigNum: os.Args[3],
			})

		case "delete":
			err = execDeleteContainerRuntime(&execInfo{
				id: containerID,
			})

		default:
			err = fmt.Errorf("Unknown command: %s\n",command)
	}
	if err != nil{
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
