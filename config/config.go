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
	"sync/atomic"

	"github.com/sirupsen/logrus"

	"github.com/wrfly/gus-proxy/proxy"
	"github.com/wrfly/gus-proxy/utils"
)

// Config ...
type Config struct {
	Debug               bool
	ProxyFilePath       string
	NoProxyCIDR         []*net.IPNet
	Scheduler           string
	ListenPort          string
	DebugPort           string
	RandomUA            bool
	ProxyUpdateInterval int
	DBFilePath          string

	proxyHostsHash      string
	proxyAliveHash      string
	proxyFilePathIsURL  bool
	proxyHosts          proxy.Hosts
	oldHosts            []string
	availableProxyHosts proxy.Hosts

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
	case proxy.ROUND_ROBIN:
	case proxy.RANDOM:
	case proxy.PING:
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
	var (
		proxyfile  io.ReadCloser
		proxyHosts proxy.Hosts
		newHosts   []string
		err        error
	)

	if c.proxyFilePathIsURL {
		resp, err := http.DefaultClient.Get(c.ProxyFilePath)
		if err != nil {
			return fmt.Errorf("get host %s error: %s", c.ProxyFilePath, err)
		}
		proxyfile = resp.Body
	} else {
		proxyfile, err = os.Open(c.ProxyFilePath)
		if err != nil {
			return err
		}
	}
	defer proxyfile.Close()
	s := bufio.NewScanner(proxyfile)
	var l int
	for s.Scan() {
		l++
		target := s.Text()
		target = strings.ToValidUTF8(target, "")
		target = strings.TrimFunc(target, func(r rune) bool {
			if r == '\r' || r == '\n' {
				return true
			}
			return false
		})
		if err := s.Err(); err != nil {
			return fmt.Errorf("read hosts error: %s", err)
		}
		if target == "" || strings.HasPrefix(target, "#") {
			// skip empty line and comments
			continue
		}

		// verify hosts
		logrus.Debugf("validate proxy format: %s", target)
		proxyline, err := url.Parse(target)
		if err != nil {
			logrus.Error(err)
			continue
		}
		if !proxyline.IsAbs() {
			logrus.Errorf("bad proxy: %s, not absolute", target)
			continue
		}

		newHosts = append(newHosts, target)
	}
	if c.proxyHostsHash == utils.HashSlice(newHosts) {
		return nil
	}
	c.proxyHostsHash = utils.HashSlice(newHosts)

	logrus.Infof("loading %d proxies", len(newHosts))

	c.m.RLock()
	oldHostsMap := make(map[string]bool, len(newHosts))
	for _, host := range c.oldHosts {
		oldHostsMap[host] = true
	}
	var (
		newProxyWG sync.WaitGroup
		limit      = make(chan struct{}, 200)
		badProxy   uint32
	)
	for i, host := range newHosts {
		if oldHostsMap[host] {
			if p := c.proxyHosts.Host(i); p != nil {
				proxyHosts.Add(p)
			}
		} else {
			newProxyWG.Add(1)
			limit <- struct{}{}
			go func(host string) {
				defer func() {
					newProxyWG.Done()
					<-limit
				}()

				proxyhost := &proxy.Host{Addr: host}
				if err := proxyhost.Init(); err != nil {
					logrus.Error(err)
					atomic.AddUint32(&badProxy, 1)
				} else {
					proxyHosts.Add(proxyhost)
				}
			}(host)
		}
	}
	newProxyWG.Wait()

	c.m.RUnlock()

	c.m.Lock()
	logrus.Warnf("load %d dead proxies", badProxy)
	logrus.Infof("load %d alive proxies", proxyHosts.Len())
	c.proxyHosts = proxyHosts
	c.m.Unlock()

	return nil
}

// UpdateProxies update proxy's attr
func (c *Config) UpdateProxies() {
	err := c.loadHosts()
	if err != nil {
		logrus.Errorf("load proxy error: %s", err)
		return
	}

	var (
		wg             sync.WaitGroup
		availableProxy int32
	)

	limit := make(chan struct{}, 1e3)
	for _, host := range c.proxyHosts.Hosts() {
		limit <- struct{}{}
		wg.Add(1)
		go func(host *proxy.Host) {
			defer wg.Done()
			if err := host.CheckAvaliable(); err != nil {
				host.Available = false
			} else {
				atomic.AddInt32(&availableProxy, 1)
			}
			logrus.Debugf("proxy: %s, Available: %t",
				host.Addr, host.Available)
			<-limit
		}(host)
	}
	wg.Wait()

	totalNum := c.proxyHosts.Len()
	// mast in this order (small to big)
	switch {
	case availableProxy*4 <= totalNum:
		logrus.Errorf("not enough available proxies, available: [%d] total: [%d]",
			availableProxy, totalNum)
	case availableProxy*2 <= totalNum:
		logrus.Warnf("half of the proxies was down, available: [%d] total: [%d]",
			availableProxy, totalNum)
	}

	oldHosts := make([]string, 0, c.proxyHosts.Len())
	for _, host := range c.proxyHosts.Hosts() {
		if host.Available {
			oldHosts = append(oldHosts, host.Addr)
		}
	}
	if c.proxyAliveHash == utils.HashSlice(oldHosts) {
		logrus.Debugf("alive proxy not changed, continue updating")
		return
	}
	c.proxyAliveHash = utils.HashSlice(oldHosts)

	c.m.Lock()
	c.availableProxyHosts = proxy.Hosts{}
	for _, ph := range c.proxyHosts.Hosts() {
		if ph.Available {
			c.availableProxyHosts.Add(ph)
		}
	}
	logrus.Infof("update %d available proxies", c.availableProxyHosts.Len())
	c.m.Unlock()

}

// ProxyHosts returns all the proxy hosts get from URL or a static file
func (c *Config) ProxyHosts() []*proxy.Host {
	return c.availableProxyHosts.Hosts()
}
