package main

import (
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


	app.Before = func(context *cli.Context) error {
		log.SetOutput(os.Stdout)

		return nil
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}