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

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/sirupsen/logrus"
	"github.com/wrfly/goproxy"

	"github.com/wrfly/gus-proxy/prox"
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
	proxyHosts         []*types.ProxyHost
	ProxyUpdate        int
	m                  sync.RWMutex
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
	case types.ROUND_ROBIN:
	case types.RANDOM:
	case types.PING:
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
func (c *Config) LoadHosts() error {
	logrus.Debug("load proxy hosts")
	proxyHosts := []*types.ProxyHost{}
	var proxyfile io.ReadCloser
	if c.ProxyFilePathIsURL {
		resp, err := http.DefaultClient.Get(c.ProxyFilePath)
		if err != nil {
			return err
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
			return err
		}

		if s[0] == '#' {
			continue
		}

		// verify hosts
		s = strings.TrimRight(s, "\n")
		logrus.Debugf("validate proxy format: %s", s)
		proxyline, err := url.Parse(s)
		if err != nil {
			logrus.Error(err)
			continue
		}
		if !proxyline.IsAbs() {
			logrus.Errorf("Proxy has a empty scheme: %s,file: %s,line %d",
				proxyline, c.ProxyFilePath, lnum)
			continue
		}

		host := &types.ProxyHost{
			Addr: s,
		}
		proxyHosts = append(proxyHosts, host)
	}

	opts := cmp.Options{
		// ignore type(http, socks5...)
		cmpopts.IgnoreFields(types.ProxyHost{}, "Type"),
		// ignore ping value, avaliable and the *goproxy
		cmpopts.IgnoreTypes(types.ProxyHost{}.Ping, true, &goproxy.ProxyHttpServer{}),
	}
	if !cmp.Equal(proxyHosts, c.proxyHosts, opts) {
		logrus.Info("Creating new proxies...")
		newHosts, err := prox.New(proxyHosts)
		if err != nil {
			logrus.Fatalf("Create proxyies error: %s", err)
		}
		c.m.Lock()
		c.proxyHosts = newHosts
		c.m.Unlock()
	}

	return nil
}

// UpdateProxies update proxy's attr
func (c *Config) UpdateProxies() {
	logrus.Debugf("loading hosts...")
	err := c.LoadHosts()
	if err != nil {
		logrus.Fatal(err)
	}

	var wg sync.WaitGroup
	availableProxy := struct {
		Num  int
		Lock sync.Mutex
	}{}

	for _, proxy := range c.proxyHosts {
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
	totalNum := len(c.proxyHosts)
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

func (c *Config) ProxyHosts() []*types.ProxyHost {
	c.m.RLock()
	defer c.m.RUnlock()
	ph := []*types.ProxyHost{}
	for _, p := range c.proxyHosts {
		if p.Available {
			ph = append(ph, p)
		}
	}
	return ph
}
