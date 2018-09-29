package main

import (
	"os/exec"
	"syscall"
	"os"
	"log"
)

func main() {
	// 指定被 fork 出来的新进程内的初始命令
	cmd := exec.Command("sh")
	// 指定进程的namespace参数
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Cloneflags: syscall.CLONE_NEWUTS | syscall.CLONE_NEWIPC | syscall.CLONE_NEWPID | syscall.CLONE_NEWNS ,
	}

	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		log.Fatal(err)
	}
}
