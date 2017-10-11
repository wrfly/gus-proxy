package round

import (
	"math/rand"
	"net/http"
	"sort"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/wrfly/goproxy"
	"github.com/wrfly/gus-proxy/db"
	"github.com/wrfly/gus-proxy/types"
	"github.com/wrfly/gus-proxy/utils"
)

const (
	ROUND_ROBIN = "round-robin"
	RANDOM      = "random"
	PING        = "ping"
)

// Proxy main structure
type Proxy struct {
	ProxyHosts []*types.ProxyHost
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
	for _, p := range p.ProxyHosts {
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
	case ROUND_ROBIN:
		rProxy = p.roundRobin()
	case RANDOM:
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
	if p.next == len(p.ProxyHosts) {
		p.next = 0
	}
	use := p.next
	rProxy := p.ProxyHosts[use]
	p.next++

	return rProxy
}

func (p *Proxy) randomProxy() *types.ProxyHost {
	availableProxy := []*types.ProxyHost{}
	for _, p := range p.ProxyHosts {
		if p.Available {
			availableProxy = append(availableProxy, p)
		}
	}

	source := rand.NewSource(time.Now().UnixNano())
	r := rand.New(source)
	use := r.Int() % len(availableProxy)
	rProxy := availableProxy[use]

	return rProxy
}

func (p *Proxy) pingProxy() *types.ProxyHost {
	availableProxy := []*types.ProxyHost{}
	for _, p := range p.ProxyHosts {
		if p.Available {
			availableProxy = append(availableProxy, p)
		}
	}

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
func New(proxyHosts []*types.ProxyHost, DNSdb *db.DNS, defaultUA string) *Proxy {
	logrus.Debugf("init proxy")
	if DNSdb == nil {
		logrus.Fatal("DNS DB is nil")
	}
	if len(proxyHosts) == 0 {
		logrus.Fatal("No available proxy to use")
	}
	return &Proxy{
		ProxyHosts: proxyHosts,
		dnsDB:      DNSdb,
		ua:         defaultUA,
	}
}
