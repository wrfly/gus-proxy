package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/wrfly/gus-proxy/config"
	"github.com/wrfly/gus-proxy/db"
	"github.com/wrfly/gus-proxy/gus"
)

func runGus(conf *config.Config) error {
	if conf.Debug {
		logrus.SetLevel(logrus.DebugLevel)
	}
	logrus.Info("starting gus-proxy")

	if err := conf.Validate(); err != nil {
		logrus.Fatalf("bad config error: %s", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	var wg sync.WaitGroup

	// update proxy status
	readyChan := make(chan interface{})
	wg.Add(1)
	go func() {
		defer wg.Done()
		logrus.Info("updating proxies")
		conf.UpdateProxies()
		readyChan <- true
		close(readyChan)

		tk := time.NewTicker(time.Second * time.Duration(conf.ProxyUpdateInterval))
		defer tk.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-tk.C:
				conf.UpdateProxies()
			}
		}
	}()
	<-readyChan

	// handle signals
	logrus.Debug("handle sigs")
	sigStop := make(chan os.Signal)
	signal.Notify(sigStop, syscall.SIGINT, syscall.SIGTERM)
	sigKill := make(chan os.Signal)
	signal.Notify(sigKill, os.Kill)

	// init db
	logrus.Debug("init dns db")
	dnsDB, err := db.New(conf.DBFilePath)
	if err != nil {
		logrus.Fatal(err)
	}
	defer dnsDB.Close()

	go func() {
		if !conf.Debug {
			return
		}
		addr := fmt.Sprintf(":%s", conf.DebugPort)
		logrus.Infof("debug is serving on %s", addr)
		http.ListenAndServe(addr, nil)
	}()

	srv := http.Server{
		Addr:    fmt.Sprintf("0.0.0.0:%s", conf.ListenPort),
		Handler: gus.New(conf, dnsDB),
	}
	go func() {
		wg.Add(1)
		defer wg.Done()
		logrus.Infof("serving on %s", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logrus.Error(err)
		}
	}()

	select {
	case <-sigStop:
		logrus.Info("about to stop")
		cancel()
		srvCtx, srvCancel := context.WithTimeout(context.Background(), time.Second*3)
		defer srvCancel()
		srv.Shutdown(srvCtx)
		quit := make(chan struct{})
		go func() {
			wg.Wait()
			quit <- struct{}{}
		}()
		defer close(quit)

		select {
		case <-sigStop:
			srvCancel()
			logrus.Warn("force quit!")
		case <-quit:
			logrus.Info("quit")
		}
	case <-sigKill:
		cancel()
		srv.Close()
	}

	logrus.Info("gus-proxy stopped")
	return nil
}
