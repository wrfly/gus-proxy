package types

import (
	"fmt"
	"net/http"
	"net/url"

	"github.com/elazarl/goproxy"
	"github.com/sirupsen/logrus"
	"golang.org/x/net/proxy"
)

func proxyHTTP(httpAddr string) (*goproxy.ProxyHttpServer, error) {
	proxyURL, err := url.Parse(httpAddr)
	if err != nil {
		return nil, err
	}
	logrus.Debugf("New HTTP proxy Host: %s, Port: %s", proxyURL.Host, proxyURL.Port())

	prox := goproxy.NewProxyHttpServer()
	prox.Tr.Proxy = http.ProxyURL(proxyURL)

	return prox, nil
}

func proxySocks5(socks5Addr string, auth proxy.Auth) (*goproxy.ProxyHttpServer, error) {

	dialer, err := proxy.SOCKS5("tcp", socks5Addr, &auth, proxy.Direct)
	if err != nil {
		return nil, err
	}

	prox := goproxy.NewProxyHttpServer()
	prox.Tr.Dial = dialer.Dial

	return prox, nil
}

func proxyDirect() *goproxy.ProxyHttpServer {
	return goproxy.NewProxyHttpServer()
}

func initGoProxy(host *ProxyHost) error {
	var (
		err         error
		hostAndPort string
	)

	host.Auth, host.Type, hostAndPort = splitURL(host.Addr)
	switch host.Type {
	case "direct":
		host.goProxy = proxyDirect()
	case "http":
		host.goProxy, err = proxyHTTP(host.Addr)
	case "socks5":
		host.goProxy, err = proxySocks5(hostAndPort, host.Auth)
	case "https":
	// TODO: add https proxy
	default:
		return fmt.Errorf("[%s]: unknown protocol %s", host.Addr, host.Type)
	}

	return err
}

func splitURL(URL string) (auth proxy.Auth, scheme, host string) {
	u, err := url.Parse(URL)
	if err != nil {
		return
	}
	scheme = u.Scheme
	host = u.Host
	if u.User != nil {
		auth.User = u.User.Username()
		auth.Password, _ = u.User.Password()
	}

	return auth, scheme, host
}
