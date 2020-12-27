package proxy

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

type Hosts struct {
	hosts []*Host
	m     sync.RWMutex
}

func (h *Hosts) Add(p *Host) {
	h.m.Lock()
	h.hosts = append(h.hosts, p)
	h.m.Unlock()
}

func (h *Hosts) Hosts() []*Host {
	return h.hosts
}

func (h *Hosts) Len() int32 {
	return int32(len(h.hosts))
}

func (h *Hosts) Host(i int) *Host {
	h.m.RLock()
	l := len(h.hosts)
	if l <= i {
		return nil
	}
	p := h.hosts[i]
	h.m.RUnlock()
	return p
}

// Host defines the proxy
type Host struct {
	Type      string  // http or socks5 or direct
	Addr      string  // 127.0.0.1:1080
	Ping      float32 // 66 ms
	Available bool
	Auth      proxy.Auth

	u     *url.URL
	proxy *goproxy.ProxyHttpServer
}

func (host *Host) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	host.proxy.ServeHTTP(w, r)
}

func (host *Host) Init() (err error) {
	logrus.Debugf("init proxy [%s]", host.Addr)
	host.u, err = url.Parse(host.Addr)
	if err != nil {
		return err
	}

	if host.u.Scheme != "direct" {
		conn, err := net.DialTimeout("tcp", host.u.Host, 3*time.Second)
		if err != nil {
			return err
		}
		conn.Close()
	}

	return initGoProxy(host)
}

func (host *Host) CheckAvaliable() (err error) {
	logrus.Debugf("check [%s] avaliable", host.Addr)

	cli := &http.Client{Timeout: 3 * time.Second}
	if host.proxy != nil {
		cli.Transport = host.proxy.Tr
	}
	defer cli.CloseIdleConnections()

	resp, err := cli.Get("http://ip.kfd.me")
	if err != nil {
		return fmt.Errorf("request failed: %s", err)
	}
	defer resp.Body.Close()

	bs, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	if len(bs) >= 1 {
		bs = bs[:len(bs)-1]
	}
	logrus.Debugf("proxy [%s] got IP %s", host.Addr, bs)
	if host.Type != DIRECT && string(bs) == localIP {
		return fmt.Errorf("bad proxy [%s], same IP", host.Addr)
	}

	host.Available = true
	// TODO: ping is not avaliable since different operating system
	// has different ping
	// host.Ping = utils.Ping(host.u.Hostname())

	return nil
}
