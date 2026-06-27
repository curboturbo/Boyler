package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"
	"log"
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

	log.Println("OPEN <signal.fifo> to write (init.go): ", pathSend)

	signalPipe, err := os.OpenFile(pathSend,os.O_WRONLY,0)
	if err != nil{
		return fmt.Errorf("Failed to open <signal.fifo:%v",err)
	}
	if err = sendSignal(signalPipe); err != nil{return err}
	defer signalPipe.Close()

	log.Println("OPEN <go.fifo> to read and wait run.go (init.go): ", pathWait)

	goPipe, err := os.OpenFile(pathWait, os.O_RDONLY,0)
	if err != nil{
		return fmt.Errorf("Failde to open <go.fifo>: %v",err)
	}
	defer goPipe.Close()

	if err = waitSignal(goPipe); err != nil{return err}
	// пересборка потоков ввода / вывода
	if err = syscall.Chroot(rootfs); err != nil{
		return fmt.Errorf("Failed chroot file system")
	}

	if err = os.Chdir("/"); err != nil {
		return fmt.Errorf("Failed chidr file system")
	}
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
    // Заставляем контейнер спать 1000 секунд
    args := []string{"/bin/sleep", "100000"}
    
    bin, err := exec.LookPath(args[0])
    if err != nil {
        return fmt.Errorf("Failed to find binary /bin/sleep: %v", err)
    }
    
    env := os.Environ()
    if err = syscall.Exec(bin, args, env); err != nil {
        return fmt.Errorf("Failed to execute /bin/sleep in container: %v", err)
    }
    return nil
}