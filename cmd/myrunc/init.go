package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"
)

func execInitContainer(i *execInfo) error {
	syscall.Sethostname([]byte(i.id))

	err := syscall.Mount("", "/", "", syscall.MS_PRIVATE|syscall.MS_REC, "") // OS cannot check containers mounting
	if err != nil {
		return fmt.Errorf("Failed isolated mount: %v", err)
	}
	rootfs := filepath.Join(i.bundlePath, "merged")
	if err = syscall.Mount(rootfs, rootfs, "", syscall.MS_BIND|syscall.MS_REC, ""); err != nil {
		return fmt.Errorf("Failed bind mount rootfs: %v", err)
	}
	procPath := filepath.Join(rootfs, "proc")
	err = os.MkdirAll(procPath, 0755)
	if err != nil {
		return fmt.Errorf("Failed create <proc> dir in container: %v", err)
	}

	err = syscall.Mount("proc", procPath, "proc", syscall.MS_NOEXEC|syscall.MS_NOSUID|syscall.MS_NODEV, "") // mount proc to container
	if err != nil {
		return fmt.Errorf("Failed mount procfs to container: %v", err)
	}

	sysPath := filepath.Join(rootfs, "sys")
	err = os.MkdirAll(sysPath, 0755)
	if err != nil {
		return fmt.Errorf("Failed create <sys> dir in container: %v", err)
	}

	err = syscall.Mount("sysfs", sysPath, "sysfs", syscall.MS_NOEXEC|syscall.MS_NOSUID|syscall.MS_NODEV|syscall.MS_RDONLY, "") // mount sys to container
	if err != nil {
		return fmt.Errorf("Failed mount sysfs to container: %v", err)
	}

	pathSend := filepath.Join(os.Getenv("STATE_PATH_MYRUNC"), i.id, os.Getenv("SIGNAL_PIPE"))
	pathWait := filepath.Join(os.Getenv("STATE_PATH_MYRUNC"), i.id, os.Getenv("GO_PIPE"))

	signalPipe, err := os.OpenFile(pathSend, os.O_WRONLY, 0)
	if err != nil {
		return fmt.Errorf("Failed to open <signal.fifo>: %v", err)
	}

	if err = sendSignal(signalPipe); err != nil {
		return err
	}
	defer signalPipe.Close()

	goPipe, err := os.OpenFile(pathWait, os.O_RDONLY, 0)
	if err != nil {
		return fmt.Errorf("Failed to open <go.fifo>: %v", err)
	}
	defer goPipe.Close()

	if err = waitSignal(goPipe); err != nil {
		return err
	}

	oldRoot := filepath.Join(rootfs, ".old_root")
	if err = os.MkdirAll(oldRoot, 0700); err != nil {
		return fmt.Errorf("Failed create .old_root dir: %v", err)
	}

	if err = syscall.PivotRoot(rootfs, oldRoot); err != nil {
		return fmt.Errorf("Failed pivot_root: %v", err)
	}

	if err = os.Chdir("/"); err != nil {
		return fmt.Errorf("Failed chdir /: %v", err)
	}

	// физически отрезаем старый (хостовый) root от mount namespace контейнера
	if err = syscall.Unmount("/.old_root", syscall.MNT_DETACH); err != nil {
		return fmt.Errorf("Failed unmount old root: %v", err)
	}

	if err = os.RemoveAll("/.old_root"); err != nil {
		fmt.Printf("warning: failed to remove /.old_root: %v\n", err)
	}
	return mockStartLinux()
}

// send signal to parent process
func sendSignal(signalPipe *os.File) error{
	_, err := signalPipe.Write(make([]byte, 1))
	if err != nil {
		return fmt.Errorf("Failed to write to <signal.fifo>: %v\n", err)
	}
	return nil
}


// wait byte sending "start" command
func waitSignal(goPipe *os.File) error {
	buffer := make([]byte, 1)
	_, err := goPipe.Read(buffer)
	if err != nil{
		return fmt.Errorf("Failed to take byte from pipe ready pipe (especial and rare fail): %v\n", err)
	}
	return nil
}


func SetupContainerDNS(rootfsPath string) error {
	resolvConfPath := filepath.Join(rootfsPath, "etc", "resolv.conf")
	etcDir := filepath.Dir(resolvConfPath)
	if err := os.MkdirAll(etcDir, 0755); err != nil {
		return fmt.Errorf("failed to create /etc directory inside container rootfs: %w", err)
	}
	dnsConfig := []byte("nameserver 8.8.8.8\nnameserver 1.1.1.1\n")
	if err := os.WriteFile(resolvConfPath, dnsConfig, 0644); err != nil {
		return fmt.Errorf("failed to write /etc/resolv.conf inside container rootfs: %w", err)
	}
	return nil
}

// "plug" before real user command for debug (onlu for linux images)
func mockStartLinux() error {
    // Заставляем контейнер спать 1000 секунд
    args := []string{"/bin/sleep", "1000"}
    
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