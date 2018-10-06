package main

import (
	"log"
	"os"
	"os/exec"
	"syscall"
)

func main() {
	// 指定被 fork 出来的新进程内的初始命令
	cmd := exec.Command("sh")

	// 指定进程的namespace参数
	// 设置启动用户用户组mapping参数
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Cloneflags:  syscall.CLONE_NEWNS | syscall.CLONE_NEWUTS | syscall.CLONE_NEWIPC | syscall.CLONE_NEWPID  | syscall.CLONE_NEWUSER | syscall.CLONE_NEWNET,
		UidMappings: []syscall.SysProcIDMap{{ContainerID: 0, HostID: 1000, Size: 1,},},
		GidMappings: []syscall.SysProcIDMap{{ContainerID: 0, HostID: 1000, Size: 1,},},
	}

	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		log.Fatal(err)
	}

	os.Exit(0)
}
