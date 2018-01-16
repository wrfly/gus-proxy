package main

import (
	"os"

	"github.com/wrfly/gus-proxy/cmds"
	"gopkg.in/urfave/cli.v2"
)

func main() {
	app := cli.App{
		Name:    "gus-proxy",
		Usage:   "An apple a day, keep the doctor away.",
		Version: "0.3",
		Authors: []*cli.Author{
			&cli.Author{
				Name:  "wrfly",
				Email: "mr.wrfly@gmail.com"},
		},
		Commands: []*cli.Command{
			cmds.Run(),
		},
	}
	app.Run(os.Args)
}
