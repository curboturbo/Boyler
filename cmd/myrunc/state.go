package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"syscall"
)


// return current state about container
func execCheckStateContainer(i *execInfo) error {
	path := filepath.Join(os.Getenv("STATE_PATH_MYRUNC"), i.id, os.Getenv("MYRUNC_META"))
	stateBytes, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("failed to read state.json: %w", err)
	}
	var containerState State
	if err := json.Unmarshal(stateBytes, &containerState); err != nil {
		return fmt.Errorf("failed to unmarshal state.json: %w", err)
	}
	proc, _ := os.FindProcess(containerState.PID)
	if err := proc.Signal(syscall.Signal(0)); err != nil {
		containerState.Status = StatusStopped
		containerDir := filepath.Join(os.Getenv("STATE_PATH_MYRUNC"), i.id)
		if err := changeState(containerDir, StatusStopped); err != nil {
			return fmt.Errorf("failed to update stopped state on disk: %w", err)
		}
	}
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(containerState); err != nil {
		return fmt.Errorf("failed to encode state to stdout: %w", err)
	}
	return nil
}
