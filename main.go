package main

import (
	"os"

	"github.com/wrfly/gus-proxy/handler"
	"gopkg.in/urfave/cli.v1"
)

func main() {
	app := cli.App{
		Name:    "gus-proxy",
		Usage:   "An apple a day, keep the doctor away.",
		Version: "0.3",
		Author:  "wrfly",
		Email:   "mr.wrfly@gmail.com",
		Commands: []cli.Command{
			handler.Run(),
		},
	}
	app.Run(os.Args)
}
