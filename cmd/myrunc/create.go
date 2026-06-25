package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"
)

// execCreateContainer fork runc process and prepare container
// before start with "execRunContainer"
func execCreateContainer(i *execInfo) error {
	runcDir := filepath.Join("/var/run/myrunc", i.id) // create root dir var temp files
	if err := os.MkdirAll(runcDir, 0755); err != nil{
		fmt.Fprintf(os.Stderr, "Failed to create state dir: %v\n", err)
		os.Exit(1)
	}

	pipePath := filepath.Join(runcDir,"signal.fifo")
	pipePath2 := filepath.Join(runcDir,"go.fifo")
	_ = os.Remove(pipePath)
	_ = os.Remove(pipePath2)

	err := syscall.Mknod(pipePath, syscall.S_IFIFO|0666, 0)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create FIFO: %v\n", err)
		os.Exit(1)
	}

	err = syscall.Mknod(pipePath2, syscall.S_IFIFO|0666, 0)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create FIFO: %v\n", err)
		os.Exit(1)
	}

	cmd := exec.Command("/proc/self/exe","init", i.id,"--bundle", i.bundlePath)
	cmd.SysProcAttr = &syscall.SysProcAttr{  // create basic isolation for childred process
		Cloneflags: syscall.CLONE_NEWPID |
		syscall.CLONE_NEWNS | 
		syscall.CLONE_NEWNET |
		syscall.CLONE_NEWUTS |
		syscall.CLONE_NEWIPC,
	}
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Env = append(os.Environ(), fmt.Sprintf("MYRUNC_BUNDLE=%s", i.bundlePath)) // bundle path presents
	// boyler/lib/containers/merged

	// fork process
	if err := cmd.Start(); err != nil{ 
		fmt.Fprintf(os.Stderr, "Failed to fork init process: %v\n", err)
		os.Exit(1)
	}
	
	readyPipe, err := os.OpenFile(pipePath, os.O_RDONLY, 0)
	if err != nil{
		fmt.Fprintf(os.Stderr, "Failed to open ready pipe: %v\n", err)
		os.Exit(1)
	}
	buffer := make([]byte, 1)
	_, err = readyPipe.Read(buffer)
	if err != nil{
		fmt.Fprintf(os.Stderr, "Failed to take byte from pipe ready pipe (especial and rare fail): %v\n", err)
		os.Exit(1)
	}
	readyPipe.Close()
	if err = createState(runcDir, i.id, cmd.Process.Pid, i.bundlePath); err != nil{
		fmt.Fprintf(os.Stderr, "Failed to save state.json, kill myself and forked process: %v\n", err)
		_ = cmd.Process.Kill()
		os.Exit(1)
	}
	return nil
}

// createState create state.json in /var/run/myrunc/contrainer_id directory
func createState(runcDir string, id string, pid int, bundlePath string) error {
	state := State{
		ID: id,
		PID: pid,
		BundlePath: bundlePath,
		Status: "created",
		OciVerion: OCI_VERSION,
	}
	
	jsonData,err := json.MarshalIndent(state, "", "    ")
	if err != nil {return err}
	path := filepath.Join(runcDir,"state.json")
	return os.WriteFile(path, jsonData, 0644)
}

type Status string

const (
	StatusCreating Status = "creating"
	StatusCreated  Status = "created"
	StatusRunning  Status = "running"
	StatusStopped  Status = "stopped"
)

const OCI_VERSION = "1.02.2"

type State struct {
	ID 			string `json:"id"`
	PID         int    `json:"pid"`
	OciVerion   string `json:"ociVersion"`
	Status      Status `json:"status"`
	BundlePath  string `json:"bundle"`
}