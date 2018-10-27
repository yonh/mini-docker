package main

import (
	"github.com/urfave/cli"
	"log"
	"os"
)

const usage  = "mini-docker is a simple container runtime implementation."

func main() {
	app := cli.NewApp()
	app.Name = "mini-docker"
	app.Usage = usage


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