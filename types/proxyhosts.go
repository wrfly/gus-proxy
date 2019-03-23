package types

import (
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"sync"
	"time"

	"github.com/elazarl/goproxy"
	"github.com/sirupsen/logrus"
	"golang.org/x/net/proxy"

	"github.com/wrfly/gus-proxy/utils"
)

var localIP string

func init() {
	ip, err := utils.PublicIP()
	if err != nil {
		panic(err)
	}
	localIP = ip
}

type ProxyHosts struct {
	hosts []*ProxyHost
	m     sync.RWMutex
}

func (h *ProxyHosts) Add(p *ProxyHost) {
	h.m.Lock()
	h.hosts = append(h.hosts, p)
	h.m.Unlock()
}

func (h *ProxyHosts) Hosts() []*ProxyHost {
	return h.hosts
}

func (h *ProxyHosts) Len() int32 {
	return int32(len(h.hosts))
}

func (h *ProxyHosts) Host(i int) *ProxyHost {
	h.m.RLock()
	l := len(h.hosts)
	if l <= i {
		return nil
	}
	p := h.hosts[i]
	h.m.RUnlock()
	return p
}

// ProxyHost defines the proxy
type ProxyHost struct {
	Type      string  // http or socks5 or direct
	Addr      string  // 127.0.0.1:1080
	Ping      float32 // 66 ms
	Available bool
	Auth      proxy.Auth

	u       *url.URL
	goProxy *goproxy.ProxyHttpServer
}

func (host *ProxyHost) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	host.goProxy.ServeHTTP(w, r)
}

func (host *ProxyHost) Init() (err error) {
	if err := host.initProxy(); err != nil {
		return err
	}
	if err := host.CheckAvaliable(); err != nil {
		return err
	}
	return nil
}

func (host *ProxyHost) initProxy() (err error) {
	logrus.Debugf("init proxy [%s]", host.Addr)
	host.u, err = url.Parse(host.Addr)

	if host.u.Scheme == "direct" {
		return initGoProxy(host)
	}

	conn, err := net.DialTimeout("tcp", host.u.Host, 1*time.Second)
	if err != nil {
		logrus.Errorf("proxy [%s] is unavailable (dial)", host.Addr)
		logrus.Debugf("error: %s", err)
		return err
	}
	conn.Close()

	return initGoProxy(host)
}

func (host *ProxyHost) CheckAvaliable() (err error) {
	logrus.Debugf("check [%s] avaliable", host.Addr)

	cli := &http.Client{Timeout: 3 * time.Second}
	if host.goProxy != nil {
		cli.Transport = host.goProxy.Tr
	}

	req, _ := http.NewRequest("GET", "http://ip.kfd.me", nil)
	resp, err := cli.Do(req)
	if err != nil {
		return fmt.Errorf("proxy [%s] is unavailable (request)", host.Addr)
	}
	defer resp.Body.Close()

	bs, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	logrus.Debugf("proxy [%s] got IP %s", host.Addr, string(bs))
	if host.Type != DIRECT && string(bs) == localIP {
		return fmt.Errorf("bad proxy [%s], same IP", host.Addr)
	}

	host.Available = true
	// TODO: ping is not avaliable since different operating system
	// has different ping
	// host.Ping = utils.Ping(host.u.Hostname())

	return nil
}
