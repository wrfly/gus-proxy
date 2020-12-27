package main

import (
	"fmt"
	"net"
	_ "net/http/pprof"
	"os"
	"sort"

	"github.com/sirupsen/logrus"
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
	conf := &config.Config{
		NoProxyCIDR: make([]*net.IPNet, 0),
	}

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
			&cli.StringSliceFlag{
				Name:    "no-proxy-cidr",
				Aliases: []string{"np"},
				Value:   cli.NewStringSlice("127.0.0.0/32"),
				Usage:   "no proxy CIDR list",
			},
			&cli.StringFlag{
				Name:        "db-path",
				Value:       "gus-proxy.db",
				Usage:       "bolt db storage",
				Destination: &conf.DBFilePath,
			},
			&cli.BoolFlag{
				Name:        "debug",
				Aliases:     []string{"d"},
				Usage:       "debug mode",
				Value:       false,
				Destination: &conf.Debug,
			},
			&cli.StringFlag{
				Name:        "scheduler",
				Aliases:     []string{"s"},
				Value:       "round_robin",
				Usage:       "scheduler: round_robin|ping|random",
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
			&cli.BoolFlag{
				Name:        "random-ua",
				Aliases:     []string{"ru"},
				Usage:       "enable random UA",
				Destination: &conf.RandomUA,
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
			logrus.SetLevel(logrus.DebugLevel)

			for _, cidr := range c.StringSlice("no-proxy-cidr") {
				_, n, err := net.ParseCIDR(cidr)
				if err != nil {
					logrus.Fatalf("invalid CIDR: %s, error: %s", cidr, err)
				}
				logrus.Debugf("append CIDR %s", n.String())
				conf.NoProxyCIDR = append(conf.NoProxyCIDR, n)
			}
			logrus.SetLevel(logrus.InfoLevel)

			return runGus(conf)
		},
	}

	sort.Sort(cli.FlagsByName(app.Flags))

	app.Run(os.Args)
}
