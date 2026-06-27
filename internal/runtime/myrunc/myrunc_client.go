package runc

import (
	"boyler/internal/runtime"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"
)

type myRunc struct {
	binaryPath string // directory with bin files {boyler/cmd}
}


func NewMyRunc(binaryPath string) runtime.Runtime {
	return &myRunc{binaryPath: binaryPath}
}


func (r *myRunc) Create(ctx context.Context, id string, bundlePath string) (*runtime.State, error) {
	absBundlePath, err := filepath.Abs(bundlePath)
	if err != nil{
		return &runtime.State{},fmt.Errorf("Invalid bundle path")
	}
	//var stdout bytes.Buffer
	cmd := exec.CommandContext(ctx, r.binaryPath, "create", id, "--bundle", absBundlePath)
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	if err = cmd.Run(); err != nil{
		return &runtime.State{}, fmt.Errorf("Failed run myrunc binary: %w", err)
	}
	fmt.Print("я закончил create")
	return &runtime.State{}, nil
}


func (r *myRunc) Run(ctx context.Context, id string) error {
	fmt.Print("я начал run")
	cmd := exec.CommandContext(ctx, r.binaryPath, "run", id, "--bundle", "absBundlePath")
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	if err := cmd.Run(); err != nil{
		return fmt.Errorf("Failed to run container: %w", err)
	}
	return nil
}


func (c *myRunc) Kill(ctx context.Context, id string, signal os.Signal) error {
	sigNum := fmt.Sprintf("%d", signal.(syscall.Signal))
	cmd := exec.CommandContext(ctx, c.binaryPath, "kill", id, sigNum)
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("myrunc kill failed: %w", err)
	}
	return nil
}


func (c *myRunc) Delete(ctx context.Context, id string) error {
	cmd := exec.CommandContext(ctx,c.binaryPath,"delete",id)
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	if err := cmd.Run(); err != nil{
		return fmt.Errorf("Failed delete for container %s: %w", id, err)
	}
	return nil
}


func (c *myRunc) State(ctx context.Context, id string) (*runtime.State, error) {
	cmd := exec.CommandContext(ctx, c.binaryPath, "state",id)
	var stdout bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil{
		return &runtime.State{}, fmt.Errorf("Failed to request state container: %w", err)
	}
	var containerState runtime.State
	if err := json.Unmarshal(stdout.Bytes(), &containerState); err != nil{
		return &runtime.State{}, fmt.Errorf("Failed to parse state.json container: %w", err)
	}
	return &containerState, nil
}