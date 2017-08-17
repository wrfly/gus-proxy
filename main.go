package main

import (
	"net"
	"net/http"
	"os"

	log "github.com/sirupsen/logrus"
	"github.com/wrfly/gus-proxy/config"
	"github.com/wrfly/gus-proxy/prox"
	"github.com/wrfly/gus-proxy/round"
	"gopkg.in/urfave/cli.v1"
)

func main() {
	var (
		debug    bool
		hostfile string
		schduler string
		listenpt string
	)
	app := cli.NewApp()
	app.Name = "gus-proxy"
	app.Usage = "An apple a day, keep the doctor away."
	app.Version = "0.0.1"

	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:        "file, f",
			Value:       "proxyhosts.txt",
			Usage:       "host list contains the proxys",
			Destination: &hostfile,
		},
		cli.BoolFlag{
			Name:        "debug, d",
			Usage:       "debug mode",
			Destination: &debug,
		},
		cli.StringFlag{
			Name:        "schduler, s",
			Value:       "round-robin",
			Usage:       "schduler: round-robin|ping|random",
			Destination: &schduler,
		},
		cli.StringFlag{
			Name:        "listen, l",
			Value:       "8080",
			Usage:       "port to bind",
			Destination: &listenpt,
		},
	}

	app.Action = func(c *cli.Context) error {
		if debug {
			log.SetLevel(log.DebugLevel)
		}
		conf := config.Config{
			ProxyHostsFile: hostfile,
			Scheduler:      schduler,
			ListenPort:     listenpt,
		}
		runGus(conf)
		return nil
	}

	app.Run(os.Args)
}

func runGus(conf config.Config) {

	if !conf.Validate() {
		log.Fatal("Config Validate Error")
	}

	hosts, err := conf.LoadHosts()
	if err != nil {
		log.Fatal(err)
	}

	proxys, err := prox.New(hosts)
	if err != nil {
		panic(err)
	}

	l, err := net.Listen("tcp4", conf.ListenPort)
	if err != nil {
		panic(err)
	}

	http.Serve(l, round.New(proxys))
}
