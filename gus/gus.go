package gus

import (
	"math/rand"
	"net/http"
	"sort"
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
	proxyHosts func() []*types.ProxyHost
	scheduler  string // round-robin/random/ping
	ua         string
	dnsDB      *db.DNS
	next       int
}

func (gs *Gustavo) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	logrus.Debugf("Request: %v", r.URL)
	defer r.Body.Close()
	// rebuild request
	r.URL.Host = utils.SelectIP(r.Host, gs.dnsDB)
	if gs.ua != "" {
		r.Header.Set("User-Agent", utils.RandomUA())
	}

	selectedProxy := gs.SelectProxy()
	if selectedProxy != nil {
		logrus.Debugf("Use proxy: %s", selectedProxy.Addr)
		selectedProxy.GoProxy.ServeHTTP(w, r)
		if w.Header().Get("PROXY_CODE") == "500" {
			selectedProxy.Available = false
			// proxy is down
			logrus.Errorf("Proxy [%s] is down", selectedProxy.Addr)
		}
		return
	}
	// when thereis no proxy available, we connect the target directly
	logrus.Error("No proxy available, direct connect")
	gp := goproxy.NewProxyHttpServer()
	gp.ServeHTTP(w, r)
}

// SelectProxy returns a proxy depends on your scheduler
func (gs *Gustavo) SelectProxy() (rProxy *types.ProxyHost) {
	// make sure that we can select one at least
	proxyavailable := false
	ph := gs.proxyHosts()
	for _, p := range ph {
		if p.Available {
			proxyavailable = true
			break
		}
	}
	if !proxyavailable {
		return nil
	}

ReSelect:
	switch gs.scheduler {
	case types.ROUND_ROBIN:
		rProxy = gs.roundRobin()
	case types.RANDOM:
		rProxy = gs.randomProxy()
	default:
		rProxy = gs.roundRobin()
	}
	if !rProxy.Available {
		goto ReSelect
	}

	return rProxy
}

func (gs *Gustavo) roundRobin() *types.ProxyHost {
	ph := gs.proxyHosts()
	if gs.next == len(ph) {
		gs.next = 0
	}
	use := gs.next
	rProxy := ph[use]
	gs.next++

	return rProxy
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
		proxyHosts: conf.ProxyHosts,
		scheduler:  conf.Scheduler,
		dnsDB:      DNSdb,
		ua:         conf.UA,
	}
}
