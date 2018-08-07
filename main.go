package main

import (
	"fmt"
	"os"

	"github.com/wrfly/gus-proxy/cmds"
	"gopkg.in/urfave/cli.v2"
)

func main() {
	app := cli.App{
		Name:  "gus-proxy",
		Usage: "An apple a day, keep the doctor away.",
		Version: fmt.Sprintf("version: %s\tcommit: %s\tdate: %s",
			Version, CommitID, BuildAt),
		Authors: author,
		Commands: []*cli.Command{
			cmds.Run(),
		},
	}
	app.Run(os.Args)
}
