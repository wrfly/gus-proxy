package handler

import (
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/wrfly/gus-proxy/config"
	"github.com/wrfly/gus-proxy/db"
	"github.com/wrfly/gus-proxy/prox"
	"github.com/wrfly/gus-proxy/round"
	cli "gopkg.in/urfave/cli.v1"
)

func Run() cli.Command {
	conf := &config.Config{}

	runFlags := []cli.Flag{
		cli.StringFlag{
			Name:        "file, f",
			Value:       "proxyhosts.txt",
			Usage:       "proxy file path, filepath or URL",
			Destination: &conf.ProxyFilePath,
		},
		cli.BoolFlag{
			Name:        "debug, d",
			Usage:       "debug mode",
			Destination: &conf.Debug,
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

	return cli.Command{
		Name:  "run",
		Usage: "Run gus-proxy",
		Flags: runFlags,
		Action: func(c *cli.Context) error {
			return runGus(conf)
		},
	}
}

func runGus(conf *config.Config) error {
	if conf.Debug {
		logrus.SetLevel(logrus.DebugLevel)
	}
	logrus.Info("Gus is starting...")

	if err := conf.Validate(); err != nil {
		logrus.Fatalf("Verify config error: %s", err)
	}

	hosts, err := conf.LoadHosts()
	if err != nil {
		logrus.Fatal(err)
	}

	logrus.Info("Creating proxys...")
	conf.ProxyHosts, err = prox.New(hosts)
	if err != nil {
		logrus.Fatalf("Create proxys error: %s", err)
	}

	// update proxy status
	logrus.Info("Updating proxys...")
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
	logrus.Debug("handle sigs")
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	// init db
	logrus.Debug("init dns db")
	DNSdb, err := db.New()
	if err != nil {
		logrus.Fatal(err)
	}
	defer DNSdb.Close()

	go func() {
		logrus.Debug("bind port and run")
		l, err := net.Listen("tcp4", conf.ListenPort)
		if err != nil {
			logrus.Fatalf("Bind port error: %s", err)
		}

		h := round.New(conf.ProxyHosts, DNSdb, conf.UA)
		logrus.Info("Gus is running...")
		logrus.Fatal(http.Serve(l, h))
	}()

	<-sigs
	logrus.Info("Gus stopped")
	return nil
}
