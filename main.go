package main

import (
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/wrfly/gus-proxy/config"
	"github.com/wrfly/gus-proxy/db"
	"github.com/wrfly/gus-proxy/prox"
	"github.com/wrfly/gus-proxy/round"
	"gopkg.in/urfave/cli.v1"
)

func main() {

	conf := &config.Config{}

	app := cli.NewApp()
	app.Name = "gus-proxy"
	app.Usage = "An apple a day, keep the doctor away."
	app.Version = "0.0.1"

	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:        "file, f",
			Value:       "proxyhosts.txt",
			Usage:       "host list contains the proxys",
			Destination: &conf.ProxyHostsFile,
		},
		cli.BoolFlag{
			Name:  "debug, d",
			Usage: "debug mode",
		},
		cli.StringFlag{
			Name:        "schduler, s",
			Value:       "round-robin",
			Usage:       "schduler: round-robin|ping|random",
			Destination: &conf.Scheduler,
		},
		cli.StringFlag{
			Name:        "listen, l",
			Value:       "8080",
			Usage:       "port to bind",
			Destination: &conf.ListenPort,
		},
		cli.StringFlag{
			Name:        "ua, u",
			Value:       "",
			Usage:       "specific UA, random UA if empty",
			Destination: &conf.UA,
		},
	}

	app.Action = func(c *cli.Context) error {
		if c.Bool("debug") {
			log.SetLevel(log.DebugLevel)
		}
		runGus(conf)
		return nil
	}

	app.Run(os.Args)
}

func runGus(conf *config.Config) {
	log.Info("Gus is starting...")

	if !conf.Validate() {
		log.Fatal("Verify config error, exit.")
	}

	hosts, err := conf.LoadHosts()
	if err != nil {
		log.Fatal(err)
	}

	log.Info("Creating proxys...")
	conf.ProxyHosts, err = prox.New(hosts)
	if err != nil {
		log.Fatalf("Create proxys error: %s", err)
	}

	// update proxy status
	log.Info("Updating proxys...")
	upChan := make(chan interface{})
	go func() {
		for {
			conf.UpdateProxys()
			upChan <- true
			time.Sleep(1000 * time.Second)
		}
	}()
	<-upChan
	close(upChan)

	// handle signals
	log.Debug("handle sigs")
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	// init db
	log.Debug("init dns db")
	DNSdb := &db.DNS{}
	if err := DNSdb.Open(); err != nil {
		log.Fatal(err)
	}
	defer DNSdb.Close()

	go func() {
		log.Debug("bind port and run")
		l, err := net.Listen("tcp4", conf.ListenPort)
		if err != nil {
			log.Fatalf("Bind port error: %s", err)
		}

		h := round.New(conf.ProxyHosts, DNSdb, conf.UA)
		log.Info("Gus is running...")
		log.Fatal(http.Serve(l, h))
	}()

	<-sigs
	log.Info("Gus stopped")
}
