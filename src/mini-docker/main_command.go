package main

import (
	"fmt"
	"github.com/urfave/cli"
	"log"
	"mini-docker/cgroup"
	"mini-docker/cgroup/subsystem"
	"mini-docker/container"
	"os"
	"strings"
)

var runCommand = cli.Command{
	Name: "run",
	Usage: "create a container with namespace and cgroup, docker run -it [command]",
	Flags: []cli.Flag {
		cli.BoolFlag{
			Name: "it",
			Usage: "enable tty",
		},
		cli.StringFlag{
			Name: "m",
			Usage: "memory limit",
		},
		cli.StringFlag{
			Name: "cpushare",
			Usage: "cpushare limit",
		},
		cli.StringFlag{
			Name: "cpuset",
			Usage: "cpuset limit",
		},
	},
	Action: func(context *cli.Context) error {
		if len(context.Args()) < 1 {
			return fmt.Errorf("Missing container command")
		}

		var cmdArray []string
		for _, arg := range context.Args() {
			cmdArray = append(cmdArray, arg)
		}
		tty := context.Bool("it")
		resConf := &subsystem.ResourceConfig{
			MemoryLimit: context.String("m"),
			CpuSet: context.String("cpuset"),
			CpuShare:context.String("cpushare"),
		}

		Run(tty, cmdArray, resConf)

		return nil
	},
}

// init方法作为容器初始化方法，作为内部方法，禁止外部调用
var initCommand = cli.Command{
	Name: "init",
	Usage: "init container process run user's process in container. Don't call it outside.",
	Action: func(context *cli.Context) error {

		log.Println("init...")
		err := container.RunContainerInitProcess()
		return err
	},
}



// src/mini-docker/main_command.go

/**
 * 这里会启动一个子进程，子进程传入init参数，使得子进程调用init实现初始化的调用
 * + 首先我们的command参数不再接收原来单一的 -it, 而是一个数组
 */
func Run(tty bool,  cmdArr []string, res *subsystem.ResourceConfig) {
	// NewParentProcess初始化进程
	// + 这里返回数据不仅仅返回parent了,同时还返回了各writePipe,这个writePipe是什么东西呢? os.Pipe()返回的writer
	// 详情可以来这里看这篇文章 [https://www.jianshu.com/p/aa207155ca7d?utm_campaign=studygolang.com&utm_medium=studygolang.com&utm_source=studygolang.com](理解golang io.Pipe)
	parent, writePipe := container.NewParentProcess(tty)
	if parent == nil {
		log.Printf("New parent process error")
		return
	}

	if err := parent.Start(); err != nil {
		log.Fatal(err)
	}

	// + 通过cgroupManager配置创建Cgroup
	cgroupManager := cgroup.NewCgroupManager("mini-docker")
	defer cgroupManager.Destroy()
	//设置资源限制
	cgroupManager.Set(res)
	// 将容器的进程加入到各各subsystem挂载的cgroup
	cgroupManager.Apply(parent.Process.Pid)

	// 初始化容器
	sendInitCommand(cmdArr, writePipe)

	parent.Wait()
	//os.Exit(-1)
}

func sendInitCommand(cmdArr []string, writePipe *os.File) {
	command := strings.Join(cmdArr, " ")
	log.Printf("command all is %s", command)
	writePipe.WriteString(command)
	writePipe.Close()
}
