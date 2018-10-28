package main

import (
	"fmt"
	"github.com/urfave/cli"
	"log"
	"os"
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
		fmt.Println("hello, I'm run command.")
		return nil
	},
}
var initCommand = cli.Command{
	Name: "init",
	Usage: "init container process run user's process in container. Don't call it outside.",
	Action: func(context *cli.Context) error {
		return nil
	},
}
