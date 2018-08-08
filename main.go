package main

import (
	"fmt"
	// go  pprof ...
	_ "net/http/pprof"
	"os"

	"gopkg.in/urfave/cli.v2"

	"github.com/wrfly/gus-proxy/config"
)

const helpTemplate = `NAME:
    {{.Name}} - {{.Usage}}
{{if len .Authors}}
AUTHOR:
    {{range .Authors}}{{ . }}{{end}}
{{end}}{{if .Version}}
VERSION:
    {{.Version}}
{{end}}{{if .Commands}}
OPTIONS:
{{range .VisibleFlags}}    {{.}}
{{end}}{{end}}`

func main() {
	conf := &config.Config{}

	app := cli.App{
		Name:  "gus-proxy",
		Usage: "Change proxy for every request",
		Version: fmt.Sprintf("version: %s\tcommit: %s\tdate: %s",
			Version, CommitID, BuildAt),
		Authors: author,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:        "file",
				Aliases:     []string{"f"},
				Value:       "proxyhosts.txt",
				Usage:       "proxy file path, filepath or URL",
				Destination: &conf.ProxyFilePath,
			},
			&cli.BoolFlag{
				Name:        "debug",
				Aliases:     []string{"d"},
				Usage:       "debug mode",
				Destination: &conf.Debug,
			},
			&cli.StringFlag{
				Name:        "schduler",
				Aliases:     []string{"s"},
				Value:       "round-robin",
				Usage:       "schduler: round-robin|ping|random",
				Destination: &conf.Scheduler,
			},
			&cli.StringFlag{
				Name:        "listen",
				Aliases:     []string{"l"},
				Value:       "8080",
				Usage:       "port to bind",
				Destination: &conf.ListenPort,
			},
			&cli.StringFlag{
				Name:        "debug-port",
				Value:       "8081",
				Usage:       "port for pprof debug",
				Destination: &conf.DebugPort,
			},
			&cli.StringFlag{
				Name:        "ua",
				Value:       "",
				Usage:       "specific UA, random UA if empty",
				Destination: &conf.UA,
			},
			&cli.IntFlag{
				Name:        "update",
				Value:       30,
				Usage:       "proxies update interval(second)",
				Destination: &conf.ProxyUpdateInterval,
			},
		},
		CustomAppHelpTemplate: helpTemplate,
		Action: func(c *cli.Context) error {
			return runGus(conf)
		},
	}
	app.Run(os.Args)
}
