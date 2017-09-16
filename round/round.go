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

// Proxy ...
type Proxy struct {
	ProxyHosts []*types.ProxyHost
	Scheduler  string // round-robin/random/ping
	ua         string
	dnsDB      *db.DNS
	next       int
}

func (p *Proxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	// rebuild request
	r.URL.Host = utils.SelectIP(r.Host, p.dnsDB)
	r.Header.Set("User-Agent", utils.SelectUA(p.ua))

	selectedProxy := p.SelectProxy()
	if selectedProxy != nil {
		selectedProxy.GoProxy.ServeHTTP(w, r)
		if w.Header().Get("PROXY_CODE") == "500" {
			selectedProxy.Available = false
			// proxy is down
			logrus.Errorf("Proxy [%s] is down", selectedProxy.Addr)
		}
		return
	}
	// when thereis no proxy avaliable, we connect the target directly
	logrus.Error("No proxy avaliable, direct connect")
	gp := goproxy.NewProxyHttpServer()
	gp.ServeHTTP(w, r)
}

// SelectProxy returns a proxy depends on your scheduler
func (p *Proxy) SelectProxy() (rProxy *types.ProxyHost) {
	// make sure that we can select one at least
	proxyAvaliable := false
	for _, p := range p.ProxyHosts {
		if p.Available {
			proxyAvaliable = true
			break
		}
	}
	if !proxyAvaliable {
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
	logrus.Debugf("Use proxy: %s", rProxy.Addr)

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
	avaliableProxy := []*types.ProxyHost{}
	for _, p := range p.ProxyHosts {
		if p.Available {
			avaliableProxy = append(avaliableProxy, p)
		}
	}

	source := rand.NewSource(time.Now().UnixNano())
	r := rand.New(source)
	use := r.Int() % len(avaliableProxy)
	rProxy := avaliableProxy[use]

	return rProxy
}

func (p *Proxy) pingProxy() *types.ProxyHost {
	avaliableProxy := []*types.ProxyHost{}
	for _, p := range p.ProxyHosts {
		if p.Available {
			avaliableProxy = append(avaliableProxy, p)
		}
	}

	sort.Slice(avaliableProxy, func(i, j int) bool {
		return avaliableProxy[i].Ping < avaliableProxy[j].Ping
	})

	source := rand.NewSource(time.Now().UnixNano())
	r := rand.New(source)
	// 随机取前1/3 || len
	use := r.Int() % (func() int {
		l := len(avaliableProxy)
		if l <= 3 {
			return l
		}
		return l / 3
	}())
	rProxy := avaliableProxy[use]

	return rProxy
}

// New round proxy servers
func New(proxyHosts []*types.ProxyHost, DNSdb *db.DNS, defaultUA string) *Proxy {
	if len(proxyHosts) == 0 {
		logrus.Fatal("No avaliable proxy to use")
	}
	return &Proxy{
		ProxyHosts: proxyHosts,
		dnsDB:      DNSdb,
		ua:         defaultUA,
	}
}
