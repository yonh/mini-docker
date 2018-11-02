package main

import (
	"fmt"
	"github.com/urfave/cli"
	"log"
	"mini-docker/container"
	"os"
)

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
		err := container.RunContainerInitProcess(cmd, nil)
		return err
	},
}




/**
 * 这里会启动一个子进程，子进程传入init参数，使得子进程调用init实现初始化的调用
 */
func Run(tty bool, command string) {
	// NewParentProcess初始化进程
	parent := container.NewParentProcess(tty, command)
	if err := parent.Start(); err != nil {
		log.Fatal(err)
	}

	parent.Wait()
	os.Exit(-1)
}
