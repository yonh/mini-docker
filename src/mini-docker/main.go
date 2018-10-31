package main

import (
	"fmt"
	"github.com/urfave/cli"
	"log"
	"os"
	"os/exec"
	"syscall"
)

func main() {
	// 创建App,初始化相关参数
	app := cli.NewApp()
	app.Name = "mini-docker"
	app.Usage = "mini-docker is a simple container runtime implementation."
	app.Version = "0.0.1"

	app.Commands = []cli.Command{
		initCommand,
		runCommand,
	}

	app.Before = func(context *cli.Context) error {
		log.SetOutput(os.Stdout)

		return nil
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}

var runCommand = cli.Command{
	Name: "run",
	Usage: "create a container with namespace and cgroup, docker run -it [command]",
	Flags: []cli.Flag {
		cli.BoolFlag{
			Name: "it",
			Usage: "enable tty",
		},
	},
	Action: func(context *cli.Context) error {
		if len(context.Args()) < 1 {
			return fmt.Errorf("Missing container command")
		}
		cmd := context.Args().Get(0)
		tty := context.Bool("it")
		Run(tty, cmd)

		return nil
	},
}
// init方法作为容器初始化方法，作为内部方法，禁止外部调用
var initCommand = cli.Command{
	Name: "init",
	Usage: "init container process run user's process in container. Don't call it outside.",
	Action: func(context *cli.Context) error {

		log.Println("init...")
		cmd := context.Args().Get(0)
		log.Printf("initCommand %s\n", cmd)
		err := RunContainerInitProcess(cmd, nil)
		return err
	},
}


/**
 * 这里会启动一个子进程，子进程传入init参数，使得子进程调用init实现初始化的调用
 */
func Run(tty bool, command string) {
	// NewParentProcess初始化进程
	parent := NewParentProcess(tty, command)
	if err := parent.Start(); err != nil {
		log.Fatal(err)
	}

	parent.Wait()
	os.Exit(-1)
}

/**
 * /proc/self/exe 指代的是当前进程本身，通过这种方式对创建出来的进程进行初始化
 * 使得新进程在 namespace 隔离中，同时如果指定tty=true的话，将当前进程的输入输出转到新进程
 * 当前command对象启动后会调用initCommand命令，并传入args参数
 */
func NewParentProcess(tty bool, command string) *exec.Cmd {
	args := []string{"init", command}
	cmd := exec.Command("/proc/self/exe", args...)
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
	return cmd
}

/**
 * 这里的代码是由容器内这个进程执行的
 * 这里首先mount proc 文件系统
 */
func RunContainerInitProcess(command string, args []string) error {
	log.Printf("RunContainerInitProcess: %s \n", command)

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
	return nil
}