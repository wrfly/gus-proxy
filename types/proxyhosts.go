package types

import (
	"net"
	"net/http"
	"net/url"
	"time"

	"github.com/elazarl/goproxy"
	"github.com/sirupsen/logrus"
	"golang.org/x/net/proxy"
)

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
		logrus.Errorf("Proxy [%s] is not available, error: %s", host.Addr, err)
		return err
	}
	conn.Close()

	return initGoProxy(host)
}

func (host *ProxyHost) CheckAvaliable() (err error) {
	logrus.Debugf("CheckProxyAvailable [%s]", host.Addr)

	clnt := &http.Client{Timeout: 3 * time.Second}
	if host.goProxy != nil {
		clnt.Transport = host.goProxy.Tr
	}

	req, _ := http.NewRequest("GET",
		"http://www.baidu.com/home/msg/data/personalcontent", nil)
	req.Header.Set("HOST", "61.135.169.125")
	_, err = clnt.Do(req)
	if err != nil {
		host.Available = false
		logrus.Debugf("Proxy [%s] is not available, error: %s", host.Addr, err)
		return err
	}
	host.Available = true
	// TODO: ping is not avaliable since different operating system
	// has different ping
	// host.Ping = utils.Ping(host.u.Hostname())

	return nil
}
