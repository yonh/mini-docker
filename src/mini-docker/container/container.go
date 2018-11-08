package container

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"strings"
	"syscall"
)

/**
 * 这里的代码是由容器内这个进程执行的
 * 这里首先mount proc 文件系统
 * + 我们去掉原本函数的参数
 */
func RunContainerInitProcess() error {
	cmdArr := readUserCommands()
	if cmdArr == nil || len(cmdArr) < 1 {
		return fmt.Errorf("Run container get user command error, Commands is nil")
	}

	setupMount()

	path, err := exec.LookPath(cmdArr[0])
	if err != nil {
		log.Printf("Exec loop path error %v", err)
		return err
	}

	log.Printf("Find path %s", path)
	if err := syscall.Exec(path, cmdArr[0:], os.Environ()); err != nil {
		log.Printf(err.Error())
	}
	return nil
}

func setupMount() {
	//log.Printf("RunContainerInitProcess: %s \n", command)

	// MS_NOEXEC 在本文件系统下不允许运行其他程序
	// MS_NOSUID 不允许set uid 或 set gid
	// MS_NODEV  这个是自linux 2.4以来，所有mount都会默认设置的一个参数
	defaultMountFlags := syscall.MS_NOEXEC | syscall.MS_NOSUID | syscall.MS_NODEV
	syscall.Mount("proc", "/proc", "proc", uintptr(defaultMountFlags), "")
	argv := []string{command}

	// 这里的syscall.Exec比较重要，按照我们目前的代码看来，容器执行后执行的第一个进程并不是我们指定的进程
	// 那我们通过ps查看到的pid=1的进程不会是我们指定的进程，通过syscall.Exec 这个方法就可以解决这个问题
	// 它会覆盖当前进程的镜像，数据，和堆栈信息，包括PID，使得我们可以替换掉本身的init进程。
	if err := syscall.Exec(command, argv, os.Environ()); err != nil {
		log.Fatal(err.Error())
	}
}

func readUserCommands() []string {
	pipe := os.NewFile(uintptr(3), "pipe")
	msg, err := ioutil.ReadAll(pipe)
	if err != nil {
		//log.Errorf("init read pipe error %v", err)
		log.Printf("[ERROR] init read pipe error %v", err)
		return nil
	}
	msgStr := string(msg)
	return strings.Split(msgStr, " ")
}

/**
 * /proc/self/exe 指代的是当前进程本身，通过这种方式对创建出来的进程进行初始化
 * 使得新进程在 namespace 隔离中，同时如果指定tty=true的话，将当前进程的输入输出转到新进程
 * 当前command对象启动后会调用initCommand命令，并传入args参数
 *
 * + 这里去掉了原来的command参数，因为已经不需要传进来参数了
 * + 同时函数增加返回值 writePipe 对象
 */
func NewParentProcess(tty bool) (*exec.Cmd, *os.File ) {
	readPipe, writePipe, err := NewPipe()

	if (err != nil) {
		log.Printf("New pipe error %v", err)
		return nil, nil
	}

	cmd := exec.Command("/proc/self/exe", "init")
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Cloneflags: syscall.CLONE_NEWUTS | syscall.CLONE_NEWPID | syscall.CLONE_NEWNS |
			syscall.CLONE_NEWNET | syscall.CLONE_NEWIPC,
	}
	// 如果用户指定了 -it 参数, 就把当前进程的输入输出导入到标准输入输出上
	if tty {
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	}
	cmd.ExtraFiles = []*os.File{readPipe}

	return cmd, writePipe
}

func NewPipe() (*os.File, *os.File, error) {
	read, write, err := os.Pipe()
	if err != nil {
		return nil, nil, err
	}
	return read, write, nil
}