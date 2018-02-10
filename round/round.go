package round

import (
	"math/rand"
	"net/http"
	"sort"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/wrfly/goproxy"
	"github.com/wrfly/gus-proxy/config"
	"github.com/wrfly/gus-proxy/db"
	"github.com/wrfly/gus-proxy/types"
	"github.com/wrfly/gus-proxy/utils"
)

// Proxy main structure
type Proxy struct {
	proxyHosts func() []*types.ProxyHost
	Scheduler  string // round-robin/random/ping
	ua         string
	dnsDB      *db.DNS
	next       int
}

func (p *Proxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	logrus.Debugf("Request: %v", r.URL)
	defer r.Body.Close()
	// rebuild request
	r.URL.Host = utils.SelectIP(r.Host, p.dnsDB)
	r.Header.Set("User-Agent", utils.SelectUA(p.ua))

	selectedProxy := p.SelectProxy()
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
func (p *Proxy) SelectProxy() (rProxy *types.ProxyHost) {
	// make sure that we can select one at least
	proxyavailable := false
	ph := p.proxyHosts()
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
	switch p.Scheduler {
	case types.ROUND_ROBIN:
		rProxy = p.roundRobin()
	case types.RANDOM:
		rProxy = p.randomProxy()
	default:
		rProxy = p.roundRobin()
	}
	if !rProxy.Available {
		goto ReSelect
	}

	return rProxy
}

func (p *Proxy) roundRobin() *types.ProxyHost {
	ph := p.proxyHosts()
	if p.next == len(ph) {
		p.next = 0
	}
	use := p.next
	rProxy := ph[use]
	p.next++

	return rProxy
}

func (p *Proxy) randomProxy() *types.ProxyHost {
	availableProxy := p.proxyHosts()

	source := rand.NewSource(time.Now().UnixNano())
	r := rand.New(source)
	use := r.Int() % len(availableProxy)
	rProxy := availableProxy[use]

	return rProxy
}

func (p *Proxy) pingProxy() *types.ProxyHost {
	availableProxy := p.proxyHosts()

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
func New(conf *config.Config, DNSdb *db.DNS) *Proxy {
	logrus.Debugf("init proxy")
	if DNSdb == nil {
		logrus.Fatal("DNS DB is nil")
	}
	if len(conf.ProxyHosts()) == 0 {
		logrus.Fatal("No available proxy to use")
	}
	return &Proxy{
		proxyHosts: conf.ProxyHosts,
		dnsDB:      DNSdb,
		ua:         conf.UA,
	}
}
