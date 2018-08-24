package gus

import (
	"math/rand"
	"net"
	"net/http"
	"sort"
	"sync"
	"time"

	"github.com/elazarl/goproxy"
	"github.com/sirupsen/logrus"

	"github.com/wrfly/gus-proxy/config"
	"github.com/wrfly/gus-proxy/db"
	"github.com/wrfly/gus-proxy/types"
	"github.com/wrfly/gus-proxy/utils"
)

type GusProxy http.Handler

// Gustavo main structure
type Gustavo struct {
	proxyHosts  func() []*types.ProxyHost
	directProxy *goproxy.ProxyHttpServer
	noProxyList []*net.IPNet

	scheduler string // round-robin/random/ping
	randomUA  bool
	dnsDB     *db.DNS
	next      int
	m         sync.Mutex
}

func (gs *Gustavo) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	logrus.Debugf("Request: %v", r.URL)
	defer r.Body.Close()

	// rebuild request
	hostIP := ""
	hostIP, r.URL.Host = gs.dnsDB.SelectIP(r.Host)
	if gs.randomUA {
		r.Header.Set("User-Agent", utils.RandomUA())
	}

	if gs.notProxyThisHost(hostIP) {
		logrus.Debugf("do not proxy this IP(%s), goto to direct", hostIP)
		gs.directProxy.ServeHTTP(w, r)
		return
	}

	selectedProxy := gs.SelectProxy()
	if selectedProxy != nil {
		logrus.Debugf("Use proxy: %s", selectedProxy.Addr)
		selectedProxy.ServeHTTP(w, r)
		if w.Header().Get("PROXY_CODE") == "500" {
			selectedProxy.Available = false
			// proxy is down
			logrus.Errorf("Proxy [%s] is down", selectedProxy.Addr)
		}
		return
	}

	// when thereis no proxy available, we connect the target directly
	logrus.Error("No proxy available, direct connect")
	gs.directProxy.ServeHTTP(w, r)
}

// SelectProxy returns a proxy depends on your scheduler
func (gs *Gustavo) SelectProxy() (rProxy *types.ProxyHost) {
	if len(gs.proxyHosts()) == 0 {
		// no proxy avaliable
		return nil
	}

	switch gs.scheduler {
	case types.ROUND_ROBIN:
		return gs.roundRobin()
	case types.RANDOM:
		return gs.randomProxy()
	default:
		return gs.roundRobin()
	}
}

func (gs *Gustavo) roundRobin() *types.ProxyHost {
	gs.m.Lock()
	defer gs.m.Unlock()

	ph := gs.proxyHosts()
	// if next greater than total num, reset to 0
	if gs.next >= len(ph) {
		gs.next = 0
	}
	gs.next++

	return ph[gs.next-1]
}

func (gs *Gustavo) randomProxy() *types.ProxyHost {
	availableProxy := gs.proxyHosts()

	source := rand.NewSource(time.Now().UnixNano())
	r := rand.New(source)
	use := r.Int() % len(availableProxy)
	rProxy := availableProxy[use]

	return rProxy
}

func (gs *Gustavo) pingProxy() *types.ProxyHost {
	availableProxy := gs.proxyHosts()

	sort.Slice(availableProxy, func(i, j int) bool {
		return availableProxy[i].Ping < availableProxy[j].Ping
	})

	source := rand.NewSource(time.Now().UnixNano())
	r := rand.New(source)
	// 随机取前1/3 || len
	use := r.Int() % (func() int {
		l := len(availableProxy)
		if l <= 3 {
			return l
		}
		return l / 3
	}())
	rProxy := availableProxy[use]

	return rProxy
}

func (gs *Gustavo) notProxyThisHost(host string) bool {
	for _, n := range gs.noProxyList {
		logrus.Debug("net: ", n.String(), net.ParseIP(host).String())
		if n.Contains(net.ParseIP(host)) {
			return true
		}
	}
	return false
}

// New round proxy servers
func New(conf *config.Config, DNSdb *db.DNS) GusProxy {
	logrus.Debugf("init proxy")
	if DNSdb == nil {
		logrus.Fatal("DNS DB is nil")
	}
	if len(conf.ProxyHosts()) == 0 {
		logrus.Fatal("No available proxy to use")
	}
	return &Gustavo{
		proxyHosts:  conf.ProxyHosts,
		scheduler:   conf.Scheduler,
		noProxyList: conf.NoProxyCIDR,
		dnsDB:       DNSdb,
		randomUA:    conf.RandomUA,
		directProxy: goproxy.NewProxyHttpServer(),
	}
}
