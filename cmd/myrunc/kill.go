package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"syscall"
	"time"
)

// default data, must be created from ENV
const SystemTiemout = 5 * time.Second
const TickerTimeout = 50 * time.Millisecond


// send signal to process using global PID
func execKillContainer(i *execInfo) error {
	path := filepath.Join(os.Getenv("STATE_PATH_MYRUNC"), i.id, os.Getenv("MYRUNC_META"))
	containerState, err := readStateFile(path)
	if err != nil{return err}
	signal := defineSignalType(i.sigNum)
	proc, _ := os.FindProcess(containerState.PID)

	if err := proc.Signal(signal); err != nil {
		if errors.Is(err, os.ErrProcessDone) || isESRCH(err){
			return nil
		}
		return fmt.Errorf("Failed send signal to proceess: %v", err)
	}

	if err := waitExit(containerState.PID); err != nil{
		return err
	}
	if err := execDeleteContainerRuntime(i); err != nil{
		return err
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

func isESRCH(err error) bool {
	var errno syscall.Errno
	return errors.As(err, &errno) && errno == syscall.ESRCH
}

func isEPERM(err error) bool {
	var errno syscall.Errno
	return errors.As(err, &errno) && errno == syscall.EPERM
}


func waitExit(pid int) error {
	deadline := time.Now().Add(SystemTiemout)
	ticker := time.NewTicker(TickerTimeout)
	defer ticker.Stop()
	for {
		if processAlive(pid) == false{
			return nil
		}
		if time.Now().After(deadline){
			return errors.New("process did not exit within timeout")
		}
		<-ticker.C
	}
}


func processAlive(pid int) bool {
	proc, err := os.FindProcess(pid)
	if err != nil{ return false }
	err = proc.Signal(syscall.Signal(0))
	if err == nil{ return true }
	if errors.Is(err, os.ErrProcessDone){
		return false
	}
	return isEPERM(err)
}