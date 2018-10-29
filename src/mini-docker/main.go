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
		log.Printf("command %s\n", cmd)
		err := RunContainerInitProcess(cmd, nil)
		return err
	},
}

func RunContainerInitProcess(command string, args []string) error {
	log.Printf("command %s \n", command)

	defaultMountFlags := syscall.MS_NOEXEC | syscall.MS_NOSUID | syscall.MS_NODEV
	syscall.Mount("proc", "/proc", "proc", uintptr(defaultMountFlags), "")
	argv := []string{command}
	if err := syscall.Exec(command, argv, os.Environ()); err != nil {
		log.Fatal(err.Error())
	}
	return nil
}

func Run(tty bool, command string) {
	parent := NewParentProcess(tty, command)
	if err := parent.Start(); err != nil {
		//log.Error(err)
		log.Fatal(err)
	}
	parent.Wait()
	os.Exit(-1)
}

func NewParentProcess(tty bool, command string) *exec.Cmd {
	args := []string{"init", command}
	cmd := exec.Command("/proc/self/exe", args...)
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Cloneflags: syscall.CLONE_NEWUTS | syscall.CLONE_NEWPID | syscall.CLONE_NEWNS |
			syscall.CLONE_NEWNET | syscall.CLONE_NEWIPC,
	}
	if tty {
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	}
	return cmd
}

