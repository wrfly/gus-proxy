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
	switch p.Scheduler {
	case ROUND_ROBIN:
		return p.roundRobin()
	case RANDOM:
		return p.randomProxy()
	default:
		return p.roundRobin()
	}
}

func (p *Proxy) roundRobin() *goproxy.ProxyHttpServer {
	use := p.Next
	rProxy := p.ProxyHosts[use]
	p.Next++
	if p.Next == len(p.ProxyHosts) {
		p.Next = 0
	}
	logrus.Debugf("Use proxy: %s", rProxy.Addr)
	return rProxy.GoProxy
}

func (p *Proxy) randomProxy() *goproxy.ProxyHttpServer {
	use := rand.Int() % len(p.ProxyHosts)
	rProxy := p.ProxyHosts[use]
	logrus.Debugf("Use proxy: %s", rProxy.Addr)
	return rProxy.GoProxy
}

// New round proxy servers
func New(proxyHosts []*types.ProxyHost) *Proxy {
	if len(proxyHosts) == 0 {
		panic("No avaliable proxy hosts to use")
	}
	return &Proxy{
		ProxyHosts: proxyHosts,
	}
}
