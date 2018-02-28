package cmds

import (
	"context"
	"fmt"
	"net"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"runtime/debug"
	"sync"
	"syscall"
	"time"

	"github.com/sirupsen/logrus"
	"gopkg.in/urfave/cli.v2"

	"github.com/wrfly/gus-proxy/config"
	"github.com/wrfly/gus-proxy/db"
	"github.com/wrfly/gus-proxy/round"
)

func Run() *cli.Command {
	conf := &config.Config{}

	runFlags := []cli.Flag{
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
			Value:       10,
			Usage:       "Proxies update interval(second)",
			Destination: &conf.ProxyUpdateInterval,
		},
	}

	return &cli.Command{
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

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	var wg sync.WaitGroup

	// update proxy status
	upChan := make(chan interface{})
	go func() {
		wg.Add(1)
		defer wg.Done()
		logrus.Info("Updating proxyies...")
		conf.UpdateProxies()
		upChan <- true
		close(upChan)

		tk := time.NewTicker(time.Second * time.Duration(conf.ProxyUpdateInterval))
		defer tk.Stop()
		for ctx.Err() == nil {
			select {
			case <-tk.C:
				conf.UpdateProxies()
			}
		}
	}()
	<-upChan

	// handle signals
	logrus.Debug("handle sigs")
	sigStop := make(chan os.Signal)
	signal.Notify(sigStop, syscall.SIGINT, syscall.SIGTERM)
	sigKill := make(chan os.Signal)
	signal.Notify(sigKill, os.Kill)

	// init db
	logrus.Debug("init dns db")
	DNSdb, err := db.New()
	if err != nil {
		logrus.Fatal(err)
	}
	defer DNSdb.Close()

	go func() {
		if !conf.Debug {
			return
		}
		addr := fmt.Sprintf(":%s", conf.DebugPort)
		logrus.Debug("Debug is running...")
		logrus.Fatal(http.ListenAndServe(addr, nil))
	}()

	logrus.Debugf("bind port [%s] and run", conf.ListenPort)
	l, err := net.Listen("tcp4", fmt.Sprintf("0.0.0.0:%s", conf.ListenPort))
	if err != nil {
		logrus.Fatalf("Bind port error: %s", err)
	}
	go func() {
		wg.Add(1)
		defer wg.Done()
		handler := round.New(conf, DNSdb)
		logrus.Info("Gus is running...")
		logrus.Fatal(http.Serve(l, handler))
	}()

	select {
	case <-sigStop:
		logrus.Info("About to stop")
		l.Close()
		cancel()
		quit := make(chan struct{})
		go func() {
			wg.Wait()
			quit <- struct{}{}
		}()
		defer close(quit)

		select {
		case <-sigStop:
			logrus.Warn("Force quit!")
			debug.FreeOSMemory()
		case <-quit:
			logrus.Info("Quit")
		}
	case <-sigKill:
		l.Close()
		cancel()
		debug.FreeOSMemory()
	}

	logrus.Info("Gus stopped")
	return nil
}
