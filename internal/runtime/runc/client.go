package runc

import (
	"boyler/internal/runtime"
	"context"
	"path/filepath"
	"fmt"
	"os"
	"os/exec"
)

type myRunc struct {
	binaryPath string
}

func NewMyRunc(binaryPath string) runtime.Runtime {
	return &myRunc{binaryPath: binaryPath}
}


func (r *myRunc) Create(ctx context.Context, id string, bundlePath string) error {
	absBundlePath, err := filepath.Abs(bundlePath)
	if err != nil{
		return fmt.Errorf("Invalid bundle path")
	}
	cmd := exec.CommandContext(ctx, r.binaryPath, "create", id, "--bundle", absBundlePath)
	cmd.Stderr = os.Stderr
	if err = cmd.Run(); err != nil{
		return fmt.Errorf("Failed run myrunc binary")
	}
	return nil
}