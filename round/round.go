package round

import (
	"math/rand"
	"net/http"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/wrfly/goproxy"
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
	Next       int
}

func (p *Proxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// rebuild request
	r.URL.Host = utils.SelectIP(r.Host)
	r.Header.Set("User-Agent", utils.RandomUA())

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
	if p.Next == len(p.ProxyHosts) {
		p.Next = 0
	}
	use := p.Next
	rProxy := p.ProxyHosts[use]
	p.Next++

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

// New round proxy servers
func New(proxyHosts []*types.ProxyHost) *Proxy {
	if len(proxyHosts) == 0 {
		logrus.Fatal("No avaliable proxy to use")
	}
	return &Proxy{
		ProxyHosts: proxyHosts,
	}
}
