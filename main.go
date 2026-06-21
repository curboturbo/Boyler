package defalut

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Использование: sudo ./mydocker run <команда>")
		return
	}

	switch os.Args[1] {
	case "run":
		run()
	case "child":
		child()
	default:
		fmt.Println("Неизвестная команда. Доступна только: run")
	}
}

func run() {
	fmt.Printf("Родитель: Запуск изолированного окружения [PID: %d]\n", os.Getpid())
	args := append([]string{"child"}, os.Args[2:]...)
	cmd := exec.Command("/proc/self/exe", args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	// ИЗОЛЯЦИЯ: Включаем пространства имен (Namespaces) ядра Linux
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Cloneflags: syscall.CLONE_NEWPID | // Изоляция процессов (свой PID 1)
			syscall.CLONE_NEWUTS | // Изоляция имени хоста
			syscall.CLONE_NEWNS, // Изоляция точек монтирования ФС
	}

	if err := cmd.Run(); err != nil {
		fmt.Printf("Родитель: Ошибка запуска контейнера: %v\n", err)
		os.Exit(1)
	}
}

func child() {
	fmt.Printf("Контейнер: Я внутри изоляции [PID: %d]\n", os.Getpid())

	syscall.Sethostname([]byte("mini-docker"))

	targetRootfs, err := filepath.Abs("../alpine_rootfs")
	if err != nil {
		fmt.Printf("Контейнер: Ошибка определения пути к rootfs: %v\n", err)
		os.Exit(1)
	}

	if err := syscall.Chroot(targetRootfs); err != nil {
		fmt.Printf("Контейнер: Ошибка Chroot в папку %s: %v\n", targetRootfs, err)
		os.Exit(1)
	}


	if err := os.Chdir("/"); err != nil {
		fmt.Printf("Контейнер: Ошибка Chdir: %v\n", err)
		os.Exit(1)
	}

	syscall.Mount("proc", "proc", "proc", 0, "")


	cmd := exec.Command(os.Args[2], os.Args[3:]...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		fmt.Printf("Контейнер: Ошибка выполнения команды %s: %v\n", os.Args[2], err)
		os.Exit(1)
	}
}