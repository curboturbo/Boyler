package main

import (
	"fmt"
	"os"
	"path/filepath"
	"syscall"
)

// send signal to process using global PID
func execKillContainer(i *execInfo) error {
	path := filepath.Join(os.Getenv("STATE_PATH_MYRUNC"), i.id, os.Getenv("MYRUNC_META"))
	containerState, err := readStateFile(path)
	if err != nil{return err}
	signal := defineSignalType(i.sigNum)
	proc, _ := os.FindProcess(containerState.PID)
	if err := proc.Signal(signal); err != nil {
		return fmt.Errorf("Failed send signal to proceess: %v", err)
	}
	return nil
}

// define signal state
func defineSignalType(sigStr string) os.Signal {
	switch sigStr {
	case "SIGKILL", "9":
		return syscall.SIGKILL
	case "SIGTERM", "15":
		return syscall.SIGTERM
	case "SIGINT", "2":
		return syscall.SIGINT
	default:
		return syscall.SIGTERM
	}
}