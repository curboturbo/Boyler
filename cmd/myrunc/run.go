package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// send byte to pipe to start contianer life-cycle
func execRunContainer(i *execInfo) error {

	path := filepath.Join(os.Getenv("STATE_PATH_MYRUNC"), i.id)
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
func changeState(runcPath string, condition Status) error {
	pathState := filepath.Join(runcPath, os.Getenv("MYRUNC_META"))

	containerState, err := readStateFile(pathState)
	if err != nil{return err}

	containerState.Status = condition
	return writeStateFile(containerState, pathState)
}

// read state.json
func readStateFile(pathState string) (state State, err error) {
	stateBytes, err := os.ReadFile(pathState)
	if err != nil {
		return State{}, fmt.Errorf("Failed to read state.json: %v\n", err)
	}
	var containerState State
	err = json.Unmarshal(stateBytes, &containerState)
	if err != nil{
		return State{}, fmt.Errorf("Failed to open state.json:%v\n",err)
	}
	return containerState, nil
}

// update state.json
func writeStateFile(containerState State, pathState string) error{
	update, err := json.MarshalIndent(containerState, "", "  ")
	if err != nil {
		return fmt.Errorf("Failed to marshal updated state: %v\n", err)
	}
	if err := os.WriteFile(pathState, update, 0644); err != nil {
		return fmt.Errorf("Failed to write updated state.json: %v\n", err)
	}
	return nil
}