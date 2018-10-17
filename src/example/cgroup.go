package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path"
	"strconv"
	"syscall"
)

// 挂载了 memory subsystem 的 hierarchy 的根目录位置
const cgroupMemoryHierarchyMount = "/sys/fs/cgroup/memory"

func main() {

	// "/proc/self/exe"指的是本身进程
	if os.Args[0] == "/proc/self/exe" {
		// 容器进程
		fmt.Printf("[container] current pid is %d\n", syscall.Getpid())

		// 指定被 fork 出来的新进程内的初始命令
		cmd := exec.Command("sh")

		// 指定进程的namespace参数
		// 设置启动用户用户组mapping参数
		cmd.SysProcAttr = &syscall.SysProcAttr{}

		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		if err := cmd.Run(); err != nil {
			log.Fatal(err)
			os.Exit(1)
		}
	} else {

		fmt.Printf("[host] current pid is %d\n", syscall.Getpid())

		cmd := exec.Command("/proc/self/exe")

		cmd.SysProcAttr = &syscall.SysProcAttr{
			Cloneflags: syscall.CLONE_NEWNS | syscall.CLONE_NEWUTS | syscall.CLONE_NEWPID,
			//UidMappings: []syscall.SysProcIDMap{{ContainerID: 0, HostID: 1000, Size: 1,},},
			//GidMappings: []syscall.SysProcIDMap{{ContainerID: 0, HostID: 1000, Size: 1,},},
		}

		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		testCgroupPath := path.Join(cgroupMemoryHierarchyMount, "test_limit_memory")
		fileTasks := path.Join(testCgroupPath, "tasks")
		fileMemoryLimit := path.Join(testCgroupPath, "memory.limit_in_bytes")

		// 在系统默认创建挂载了memory subsystem 的 Hierarchy 上创建 cgroup
		os.Mkdir(testCgroupPath , 0755)
		// 将容器进程加入到这个 cgroup 中
		ioutil.WriteFile(fileTasks, []byte(strconv.Itoa(syscall.Getpid())), 0644)
		// 限制内存
		ioutil.WriteFile(fileMemoryLimit, []byte("100m"), 0644)

		if err := cmd.Run(); err != nil {
			log.Fatal(err)
			os.Exit(1)
		}

		cmd.Process.Wait()
	}
}
