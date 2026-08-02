package runc

import (
	"boyler/internal/runtime"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"
	"strconv"

	"github.com/creack/pty"
)

type myRunc struct {
	binaryPath string // directory with bin files {boyler/cmd}
}

func NewMyRunc(binaryPath string) runtime.Runtime {
	return &myRunc{binaryPath: binaryPath}
}

func (r *myRunc) Create(ctx context.Context, id string, bundlePath string) error {
	absBundlePath, err := filepath.Abs(bundlePath)
	if err != nil {
		return fmt.Errorf("Invalid bundle path")
	}
	cmd := exec.CommandContext(ctx, r.binaryPath, "create", id, "--bundle", absBundlePath)
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	if err = cmd.Run(); err != nil {
		return fmt.Errorf("Failed run myrunc binary: %w", err)
	}
	return nil
}

func (r *myRunc) Run(ctx context.Context, id string) error {
	cmd := exec.CommandContext(ctx, r.binaryPath, "run", id, "--bundle", "absBundlePath")
	cmd.Stderr = os.Stderr
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("Failed to run container: %w", err)
	}

	if err := cmd.Wait(); err != nil {
		return fmt.Errorf("Unknown error during execution: %w", err)
	}
	return nil
}

func (c *myRunc) Kill(ctx context.Context, id string, signal os.Signal) error {
	sigNum := fmt.Sprintf("%d", signal.(syscall.Signal))
	cmd := exec.CommandContext(ctx, c.binaryPath, "kill", id, sigNum)
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("myrunc kill failed: %w", err)
	}
	return nil
}

func (c *myRunc) Delete(ctx context.Context, id string) error {
	cmd := exec.CommandContext(ctx, c.binaryPath, "delete", id)
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("Failed delete for container %s: %w", id, err)
	}
	return nil
}

func (c *myRunc) State(ctx context.Context, id string) (*runtime.State, error) {
	cmd := exec.CommandContext(ctx, c.binaryPath, "state", id)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return &runtime.State{}, fmt.Errorf("Failed to create StdoutPipe: %w", err)
	}
	cmd.Stderr = os.Stderr
	if err := cmd.Start(); err != nil {
		return &runtime.State{}, fmt.Errorf("Failed to request state container: %w", err)
	}
	var containerState runtime.State
	if err := json.NewDecoder(stdout).Decode(&containerState); err != nil {
		return &runtime.State{}, fmt.Errorf("Failed to decode struct: %w", err)
	}

	if err := cmd.Wait(); err != nil {
		return &runtime.State{}, fmt.Errorf("Unknown error during execution: %w", err)
	}
	return &containerState, nil
}


type ptyProcess struct {
	io.ReadWriteCloser
	cmd *exec.Cmd
}

func (p *ptyProcess) Close() error {
	err := p.ReadWriteCloser.Close()
	_ = p.cmd.Wait() 
	return err
}


func (r *myRunc) ExecPTY(ctx context.Context, pid int64) (io.ReadWriteCloser, error) {
	cmd := exec.CommandContext(
		ctx,
		"sudo",
		"nsenter",
		"-t", strconv.FormatInt(pid, 10),
		"-m",
		"-u",
		"-i",
		"-n",
		"-p",
		"-r",
		"-w",
		"/bin/sh",
	)
	file, err := pty.Start(cmd)
	if err != nil {
		return nil, err
	}
	return &ptyProcess{
		ReadWriteCloser: file,
		cmd:             cmd,
	},  nil
}