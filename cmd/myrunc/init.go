package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"
)

// execInitContainer prepare container ro run
func execInitContainer(i *execInfo) error {
	syscall.Sethostname([]byte(i.id))

	err := syscall.Mount("", "/", "", syscall.MS_PRIVATE | syscall.MS_REC, "") // OS cannot check containers mounting
	if err != nil{
		fmt.Fprintf(os.Stderr, "Failed isolated mount")
		os.Exit(1)
	}

	rootfs := filepath.Join(i.bundlePath, "merged")

	procPath := filepath.Join(rootfs, "proc")
	err = os.MkdirAll(procPath, 0755)
	if err != nil{
		fmt.Fprintf(os.Stderr, "Failed create <proc> dir in container")
		os.Exit(1)
	}

	err = syscall.Mount("proc", procPath, "proc", syscall.MS_NOEXEC|syscall.MS_NOSUID|syscall.MS_NODEV, "") // mount proc to container
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed mount procfs to container")
		os.Exit(1)
	}

	sysPath := filepath.Join(rootfs, "sys")
	err = os.MkdirAll(sysPath, 0755)
	if err != nil{
		fmt.Fprintf(os.Stderr, "Failed create <sys> dir in container")
		os.Exit(1)
	}

	err = syscall.Mount("sysfs", sysPath, "sysfs", syscall.MS_NOEXEC|syscall.MS_NOSUID|syscall.MS_NODEV|syscall.MS_RDONLY, "") // mount sys ti container
	if err != nil{
		fmt.Fprintf(os.Stderr, "Failed mount sysfs to container")
		os.Exit(1)
	}

	pathSend := filepath.Join("/var/run/myrunc", i.id, "signal.fifo")
	pathWait := filepath.Join("/var/run/myrunc", i.id, "go.fifo")

	signalPipe, err := os.OpenFile(pathSend,os.O_WRONLY,0)
	if err != nil{
		fmt.Fprintf(os.Stderr,"Failed to open <signal.fifo")
		os.Exit(1)
	}

	goPipe, err := os.OpenFile(pathWait,os.O_RDONLY,0)
	if err != nil{
		fmt.Fprintf(os.Stderr, "Failde to open <go.fifo>")
		os.Exit(1)
	}

	if err = syscall.Chroot(rootfs); err != nil{
		fmt.Fprintf(os.Stderr,"Failed chroot file system")
		os.Exit(1)
	}

	if err = os.Chdir("/"); err != nil {
		fmt.Fprintf(os.Stderr, "Failed chidr file system")
		os.Exit(1)
	}
	sendSignal(signalPipe)
	waitSignal(goPipe)
	return mockStart()
}

// sendSignal send signal to parent process
func sendSignal(signalPipe *os.File) error{
	defer signalPipe.Close()
	_, err := signalPipe.Write(make([]byte, 1))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to write to <signal.fifo>: %v\n", err)
		os.Exit(1)
	}
	return nil
}
// waitSignal wait byte sending "start" command
func waitSignal(goPipe *os.File) error {
	defer goPipe.Close()
	buffer := make([]byte, 1)
	_, err := goPipe.Read(buffer)

	if err != nil{
		fmt.Fprintf(os.Stderr, "Failed to take byte from pipe ready pipe (especial and rare fail): %v\n", err)
		os.Exit(1)
	}
	return nil
}


// mockStart "plug" before real user command for debug
func mockStart() error {
	args := []string{"/bin/sh"}
	bin, err := exec.LookPath(args[0])
	if err != nil{
		fmt.Fprintf(os.Stderr, "Failed to find binary /bin/sh: %v\n", err)
		os.Exit(1)
	}
	env := os.Environ()
	if err = syscall.Exec(bin, args, env); err != nil{
		fmt.Fprintf(os.Stderr, "Failed to execute /bin/sh in container: %v\n", err)
		os.Exit(1)
	}
	return nil
}