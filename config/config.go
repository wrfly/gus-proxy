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

	"github.com/elazarl/goproxy"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/sirupsen/logrus"

	"github.com/wrfly/gus-proxy/types"
)

// Config ...
type Config struct {
	Debug               bool
	ProxyFilePath       string
	Scheduler           string
	ListenPort          string
	DebugPort           string
	UA                  string
	ProxyUpdateInterval int

	proxyFilePathIsURL  bool
	proxyHosts          []*types.ProxyHost
	availableProxyHosts []*types.ProxyHost

	m sync.RWMutex
}

// Validate the config
func (c *Config) Validate() error {
	logrus.Debugf("get proxyfile [%s]", c.ProxyFilePath)
	_, err := os.Open(c.ProxyFilePath)
	if err != nil && os.IsNotExist(err) {
		resp, err := http.DefaultClient.Get(c.ProxyFilePath)
		if err != nil {
			return fmt.Errorf("get host %s error: %s", c.ProxyFilePath, err)
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			return fmt.Errorf("Hostfile [%s] not found", c.ProxyFilePath)
		}
		c.proxyFilePathIsURL = true
	}

	switch c.Scheduler {
	case types.ROUND_ROBIN:
	case types.RANDOM:
	case types.PING:
	default:
		return fmt.Errorf("Unknown scheduler: %s", c.Scheduler)
	}

	// listen port
	l, err := net.Listen("tcp4", fmt.Sprintf("0.0.0.0:%s", c.ListenPort))
	if err != nil {
		return fmt.Errorf("Bind port error: %s", err)
	}
	l.Close()

	logrus.Debug("validate ok")
	return nil
}

// LoadHosts returns the proxy hosts
func (c *Config) loadHosts() error {
	logrus.Debug("load proxy hosts")
	var proxyfile io.ReadCloser
	proxyHosts := []*types.ProxyHost{}

	if c.proxyFilePathIsURL {
		resp, err := http.DefaultClient.Get(c.ProxyFilePath)
		if err != nil {
			return fmt.Errorf("get host %s error: %s", c.ProxyFilePath, err)
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
			return fmt.Errorf("read line error: %s", err)
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
		// ignore ping value, available and the *goproxy
		cmpopts.IgnoreTypes(types.ProxyHost{}.Ping, true, &goproxy.ProxyHttpServer{}),
	}
	if !cmp.Equal(proxyHosts, c.proxyHosts, opts) {
		logrus.Info("Creating new proxies...")
		for _, host := range proxyHosts {
			if err := host.Init(); err != nil {
				logrus.Errorf("Create proxyies error: %s", err)
				continue
			}
		}
		c.proxyHosts = proxyHosts
	}

	return nil
}

// UpdateProxies update proxy's attr
func (c *Config) UpdateProxies() {
	err := c.loadHosts()
	if err != nil {
		logrus.Fatalf("load proxy error: %s", err)
	}

	var wg sync.WaitGroup
	availableProxy := struct {
		n int
		m sync.Mutex
	}{}

	for _, proxy := range c.proxyHosts {
		wg.Add(1)
		go func(proxy *types.ProxyHost) {
			defer wg.Done()
			proxy.Available = false
			if proxy.CheckAvaliable() == nil {
				availableProxy.m.Lock()
				availableProxy.n++
				availableProxy.m.Unlock()
				proxy.Available = true
			}
			logrus.Debugf("Proxy: %s, Available: %v, Ping: %f",
				proxy.Addr, proxy.Available, proxy.Ping)
		}(proxy)
	}
	wg.Wait()

	availableNum := availableProxy.n
	totalNum := len(c.proxyHosts)
	switch { // mast in this order (small to big)
	case availableNum*4 <= totalNum:
		logrus.Errorf("Not enough available proxys, available: [%d] total: [%d]",
			availableNum, totalNum)
		// some alert
	case availableNum*2 <= totalNum:
		logrus.Warnf("Half of the proxys was down, available: [%d] total: [%d]",
			availableNum, totalNum)
	}

	c.m.Lock()
	c.availableProxyHosts = nil
	for _, ph := range c.proxyHosts {
		if ph.Available {
			c.availableProxyHosts = append(c.availableProxyHosts, ph)
		}
	}
	logrus.Debugf("append %d available proxies", len(c.availableProxyHosts))
	c.m.Unlock()

}

// ProxyHosts returns all the proxy hosts get from URL or a static file
func (c *Config) ProxyHosts() []*types.ProxyHost {
	c.m.RLock()
	defer c.m.RUnlock()
	return c.availableProxyHosts
}
