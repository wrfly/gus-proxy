package round

import (
	"math/rand"
	"net/http"

	"github.com/sirupsen/logrus"
	"github.com/wrfly/goproxy"
	"github.com/wrfly/gus-proxy/types"
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
	p.SelectProxy().ServeHTTP(w, r)
}

// SelectProxy returns a proxy depends on your scheduler
func (p *Proxy) SelectProxy() *goproxy.ProxyHttpServer {
	rProxy := &types.ProxyHost{}

	switch p.Scheduler {
	case ROUND_ROBIN:
		rProxy = p.roundRobin()
	case RANDOM:
		rProxy = p.randomProxy()
	default:
		rProxy = p.roundRobin()
	}
	if !rProxy.Available {
		return p.SelectProxy()
	}
	logrus.Debugf("Use proxy: %s", rProxy.Addr)
	return rProxy.GoProxy
}

func (p *Proxy) roundRobin() *types.ProxyHost {
	use := p.Next
	rProxy := p.ProxyHosts[use]
	p.Next++
	if p.Next == len(p.ProxyHosts) {
		p.Next = 0
	}

	return rProxy
}

func (p *Proxy) randomProxy() *types.ProxyHost {
	use := rand.Int() % len(p.ProxyHosts)
	rProxy := p.ProxyHosts[use]

	return rProxy
}

// New round proxy servers
func New(proxyHosts []*types.ProxyHost) *Proxy {
	if len(proxyHosts) == 0 {
		logrus.Fatal("No avaliable proxy hosts to use")
	}
	return &Proxy{
		ProxyHosts: proxyHosts,
	}
}
