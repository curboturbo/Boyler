package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"
)

// ИЗМЕНИТЬ ПОТОКИ ВВОДА-ВЫВОДА ДЛЯ КОНТЕЙНЕРА ОТДЕЛЬНО
// execInitContainer prepare container ro run
func execInitContainer(i *execInfo) error {
	syscall.Sethostname([]byte(i.id))

	err := syscall.Mount("", "/", "", syscall.MS_PRIVATE | syscall.MS_REC, "") // OS cannot check containers mounting
	if err != nil{
		return fmt.Errorf("Failed isolated mount: %v",err)
	}

	rootfs := filepath.Join(i.bundlePath, "merged")

	procPath := filepath.Join(rootfs, "proc")
	err = os.MkdirAll(procPath, 0755)
	if err != nil{
		return fmt.Errorf("Failed create <proc> dir in container: %v",err)
	}

	err = syscall.Mount("proc", procPath, "proc", syscall.MS_NOEXEC|syscall.MS_NOSUID|syscall.MS_NODEV, "") // mount proc to container
	if err != nil {
		return fmt.Errorf("Failed mount procfs to container: %v", err)
	}

	sysPath := filepath.Join(rootfs, "sys")
	err = os.MkdirAll(sysPath, 0755)
	if err != nil{
		return fmt.Errorf("Failed create <sys> dir in container: %v",err)
	}

	err = syscall.Mount("sysfs", sysPath, "sysfs", syscall.MS_NOEXEC|syscall.MS_NOSUID|syscall.MS_NODEV|syscall.MS_RDONLY, "") // mount sys ti container
	if err != nil{
		return fmt.Errorf("Failed mount sysfs to container: %v",err)
	}

	pathSend := filepath.Join("/var/run/myrunc", i.id, "signal.fifo")
	pathWait := filepath.Join("/var/run/myrunc", i.id, "go.fifo")

	signalPipe, err := os.OpenFile(pathSend,os.O_WRONLY,0)
	if err != nil{
		return fmt.Errorf("Failed to open <signal.fifo:%v",err)
	}
	defer signalPipe.Close()

	goPipe, err := os.OpenFile(pathWait,os.O_RDONLY,0)
	if err != nil{
		return fmt.Errorf("Failde to open <go.fifo>: %v",err)
	}
	defer goPipe.Close()

	if err = syscall.Chroot(rootfs); err != nil{
		return fmt.Errorf("Failed chroot file system")
	}

	if err = os.Chdir("/"); err != nil {
		return fmt.Errorf("Failed chidr file system")
	}
	if err = sendSignal(signalPipe); err != nil{return err}
	if err = waitSignal(goPipe); err != nil{return err}
	// пересборка потоков ввода / вывода
	return mockStart()
}


// sendSignal send signal to parent process
func sendSignal(signalPipe *os.File) error{
	_, err := signalPipe.Write(make([]byte, 1))
	if err != nil {
		return fmt.Errorf("Failed to write to <signal.fifo>: %v\n", err)
	}
	return nil
}


// waitSignal wait byte sending "start" command
func waitSignal(goPipe *os.File) error {
	buffer := make([]byte, 1)
	_, err := goPipe.Read(buffer)
	if err != nil{
		return fmt.Errorf("Failed to take byte from pipe ready pipe (especial and rare fail): %v\n", err)
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