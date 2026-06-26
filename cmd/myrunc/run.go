package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// execRunContainer send byte to pipe to stat contianer life-cycle
func execRunContainer(i *execInfo) error {
	path := filepath.Join("/var/run/myrunc", i.id)
	pipePath := filepath.Join(path, "go.fifo")
	writePipe, err := os.OpenFile(pipePath, os.O_WRONLY,0)

	if err != nil{
		return fmt.Errorf("Failed to open <go.fifo>:%v\n",err)
	}

	defer writePipe.Close()
	_, err = writePipe.Write(make([]byte, 1))
	if err != nil{
		return fmt.Errorf("Failed to write <go.fifo>: %v\n",err)
	}
	return changeState(path, "running")
}


// changeState change status <created> -> <running>
func changeState(servicePath string, condition Status) error {
	path := filepath.Join(servicePath,"state.json")

	stateBytes, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("Failed to read state.json: %v\n", err)
	}
	var containerState State
	err = json.Unmarshal(stateBytes, &containerState)
	if err != nil{
		return fmt.Errorf("Failed to open state.json:%v\n",err)
	}
	containerState.Status = condition

	update, err := json.MarshalIndent(containerState, "", "  ")
	if err != nil {
		return fmt.Errorf("Failed to marshal updated state: %v\n", err)
	}
	if err := os.WriteFile(path, update, 0644); err != nil {
		return fmt.Errorf("Failed to write updated state.json: %v\n", err)
	}
	return nil
}