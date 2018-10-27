package main

import (
	"fmt"
	"github.com/urfave/cli"
	"log"
	"mini-docker/container"
)

var runCommand = cli.Command{
	Name: "run",
	Usage: `Create a container with namespace and cgroups limit mini-docker run -it [command]`,
	Flags: []cli.Flag{
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
		tty:= context.Bool("it")
		Run(tty, cmd)

		return nil
	},
}


var initCommand = cli.Command{
	Name: "init",
	Usage: "init container process run user's process in container. Don't call it outside.",
	Action: func(context *cli.Context) error {
		log.Printf("init container...\n")
		cmd := context.Args().Get(0)
		log.Printf("command is %v\n", cmd)
		err := container.RunContainerInitProcess(cmd, nil)

		return err

	},
}