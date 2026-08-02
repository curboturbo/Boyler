package main

import (
	"fmt"
	"os"
	"path/filepath"
)

// delete runtime files (containerID/{state.json, .fifo})
func execDeleteContainerRuntime(i *execInfo) error {
	if err := os.RemoveAll(filepath.Join(os.Getenv("STATE_PATH_MYRUNC"), i.id)); err != nil{
		return fmt.Errorf("Failed to remove runtime files: %v", err)
	}
	return nil
}