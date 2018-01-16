package config

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"

	"github.com/sirupsen/logrus"
	"github.com/wrfly/gus-proxy/round"
	"github.com/wrfly/gus-proxy/types"
	"github.com/wrfly/gus-proxy/utils"
)

// Config ...
type Config struct {
	Debug              bool
	ProxyFilePath      string
	ProxyFilePathIsURL bool
	Scheduler          string
	ListenPort         string
	DebugPort          string
	UA                 string
	ProxyHosts         []*types.ProxyHost
}

// Validate the config
func (c *Config) Validate() error {
	logrus.Debugf("get proxyfile [%s]", c.ProxyFilePath)
	_, err := os.Open(c.ProxyFilePath)
	if err != nil && os.IsNotExist(err) {
		resp, err := http.DefaultClient.Get(c.ProxyFilePath)
		if err != nil {
			return err
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			return fmt.Errorf("Hostfile [%s] not exist", c.ProxyFilePath)
		}
		c.ProxyFilePathIsURL = true
	}

	switch c.Scheduler {
	case round.ROUND_ROBIN:
	case round.RANDOM:
	case round.PING:
	default:
		return fmt.Errorf("Unknown scheduler: %s", c.Scheduler)
	}

	// listen port
	c.ListenPort = fmt.Sprintf(":%s", c.ListenPort)
	l, err := net.Listen("tcp4", c.ListenPort)
	if err != nil {
		return fmt.Errorf("Can not bind this port: %s", err)
	}
	defer l.Close()

	logrus.Debug("validate ok")
	return nil
}

// LoadHosts returns the proxy hosts
func (c *Config) LoadHosts() ([]*types.ProxyHost, error) {
	logrus.Debug("load proxy hosts")
	proxyHosts := []*types.ProxyHost{}
	var proxyfile io.ReadCloser
	if c.ProxyFilePathIsURL {
		resp, err := http.DefaultClient.Get(c.ProxyFilePath)
		if err != nil {
			return nil, err
		}
		proxyfile = resp.Body
	} else {
		proxyfile, _ = os.Open(c.ProxyFilePath)
	}
	defer proxyfile.Close()
	lines := bufio.NewReader(proxyfile)
	var lnum int
	for {
		lnum++
		s, err := lines.ReadString('\n')
		if err != nil && s == "" {
			if err == io.EOF {
				break
			}
			return nil, err
		}

		if s[0] == '#' {
			continue
		}

		// verify hosts
		s = strings.TrimRight(s, "\n")
		logrus.Debugf("validate proxy format: %s", s)
		proxyline, err := url.Parse(s)
		if err != nil {
			return nil, err
		}
		if !proxyline.IsAbs() {
			return nil, fmt.Errorf("Proxy has a empty scheme: %s,file: %s,line %d",
				proxyline, c.ProxyFilePath, lnum)
		}

		host := &types.ProxyHost{
			Addr: s,
		}
		proxyHosts = append(proxyHosts, host)
	}

	c.ProxyHosts = proxyHosts
	return proxyHosts, nil
}

// UpdateProxys update proxy's attr
func (c *Config) UpdateProxys() {
	var wg sync.WaitGroup
	availableProxy := struct {
		Num  int
		Lock sync.Mutex
	}{}

	for _, proxy := range c.ProxyHosts {
		wg.Add(1)
		go func(proxy *types.ProxyHost) {
			defer wg.Done()
			proxy.Available = false
			if utils.CheckProxyAvailable(proxy) == nil {
				availableProxy.Lock.Lock()
				availableProxy.Num++
				availableProxy.Lock.Unlock()
				proxy.Available = true
			}
			proxy.Ping = utils.GetProxyPing(proxy)
			logrus.Debugf("Proxy: %s, Available: %v, Ping: %f",
				proxy.Addr, proxy.Available, proxy.Ping)
		}(proxy)
	}
	wg.Wait()

	availableNum := availableProxy.Num
	totalNum := len(c.ProxyHosts)
	switch { // mast in this order (small to big)
	case availableNum*4 <= totalNum:
		logrus.Errorf("Not enough available proxys, available: [%d] total: [%d], I'm angry!",
			availableNum, totalNum)
		// some alert
	case availableNum*2 <= totalNum:
		logrus.Warnf("Half of the proxys was down, available: [%d] total: [%d], I'm worried...",
			availableNum, totalNum)
	}
}
