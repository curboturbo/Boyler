package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"
	"log"
)

// execCreateContainer fork runc process and prepare container
// before start with "execRunContainer"
func execCreateContainer(i *execInfo) error {
	runcDir := filepath.Join("/var/run/myrunc", i.id) // create root dir var temp files
	if err := os.MkdirAll(runcDir, 0755); err != nil{
		return fmt.Errorf("Failed to create state dir: %v\n", err)
	}

	pipePath := filepath.Join(runcDir,"signal.fifo")
	pipePath2 := filepath.Join(runcDir,"go.fifo")
	_ = os.Remove(pipePath)
	_ = os.Remove(pipePath2)

	err := syscall.Mknod(pipePath, syscall.S_IFIFO|0666, 0)
	if err != nil {
		return fmt.Errorf("Failed to create FIFO: %v\n", err)
	}

	err = syscall.Mknod(pipePath2, syscall.S_IFIFO|0666, 0)
	if err != nil {
		return fmt.Errorf("Failed to create FIFO: %v\n", err)
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
		return fmt.Errorf("Failed to fork init process: %v\n", err)
	}
	log.Println("OPEN <signal.fifo> to Read (create.go): ", pipePath)
	readyPipe, err := os.OpenFile(pipePath, os.O_RDONLY, 0)
	if err != nil{
		return fmt.Errorf("Failed to open ready pipe: %v\n", err)
	}
	buffer := make([]byte, 1)
	_, err = readyPipe.Read(buffer)
	if err != nil{
		return fmt.Errorf("Failed to take byte from pipe ready pipe (especial and rare fail): %v\n", err)
	}
	readyPipe.Close()
	if err = createState(runcDir, i.id, cmd.Process.Pid, i.bundlePath); err != nil{
		_ = cmd.Process.Kill()
		return fmt.Errorf("Failed to save state.json, kill myself and forked process: %v\n", err)
	}
	log.Println("FINISH MY WORK")
	return nil
}

// createState create state.json in /var/run/myrunc/contrainer_id directory
func createState(runcDir string, id string, pid int, bundlePath string) error {
	state := State{
		ID: id,
		PID: pid,
		BundlePath: bundlePath,
		Status: "created",
		OciVersion: OCI_VERSION,
	}
	jsonData,err := json.MarshalIndent(state, "", "    ")
	if err != nil {return err}
	path := filepath.Join(runcDir,"state.json")
	os.Stdout.Write(jsonData)
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
	OciVersion   string `json:"ociVersion"`
	Status      Status `json:"status"`
	BundlePath  string `json:"bundle"`
}