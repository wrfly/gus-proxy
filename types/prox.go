package types

import (
	"encoding/base64"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/elazarl/goproxy"
	"github.com/sirupsen/logrus"
	"golang.org/x/net/proxy"

	// socks4 proxy
	_ "github.com/wrfly/gus-proxy/pkg/go-socks4"
)

const (
	ProxyAuthHeader = "Proxy-Authorization"
)

func setBasicAuth(username, password string, req *http.Request) {
	req.Header.Set(ProxyAuthHeader, fmt.Sprintf("Basic %s", basicAuth(username, password)))
}

func basicAuth(username, password string) string {
	return base64.StdEncoding.EncodeToString([]byte(username + ":" + password))
}

func proxyHTTP(httpAddr string) (*goproxy.ProxyHttpServer, error) {
	proxyURL, err := url.Parse(httpAddr)
	if err != nil {
		return nil, err
	}
	logrus.Debugf("new HTTP proxy Host: %s, Port: %s", proxyURL.Host, proxyURL.Port())

	proxyServer := goproxy.NewProxyHttpServer()
	proxyServer.Tr.Proxy = http.ProxyURL(proxyURL)
	proxyServer.OnRequest().HandleConnect(goproxy.AlwaysMitm)

	if proxyURL.User.String() != "" {
		pass, _ := proxyURL.User.Password()
		user := proxyURL.User.Username()
		proxyServer.ConnectDial = proxyServer.
			NewConnectDialToProxyWithHandler(
				proxyURL.String(),
				func(req *http.Request) {
					setBasicAuth(user, pass, req)
				},
			)
		proxyServer.OnRequest().DoFunc(func(req *http.Request,
			ctx *goproxy.ProxyCtx) (*http.Request, *http.Response) {
			setBasicAuth(user, pass, req)
			return req, nil
		})
	}

	return proxyServer, nil
}

func proxySocks4(u *url.URL) (*goproxy.ProxyHttpServer, error) {
	dialer, err := proxy.FromURL(u, proxy.Direct)
	if err != nil {
		return nil, err
	}

	prox := goproxy.NewProxyHttpServer()
	prox.Tr.Dial = dialer.Dial

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

	var u *url.URL
	u, host.Auth, host.Type, hostAndPort = splitURL(host.Addr)
	switch host.Type {
	case DIRECT:
		host.goProxy = proxyDirect()
	case HTTP:
		host.goProxy, err = proxyHTTP(host.Addr)
	case SOCKS5:
		host.goProxy, err = proxySocks5(hostAndPort, host.Auth)
	case HTTPS:
		// TODO:
		// host.goProxy, err = proxyHTTPS(host.Addr)
	case SOCKS4:
		host.goProxy, err = proxySocks4(u)
	default:
		return fmt.Errorf("[%s]: unknown protocol %s", host.Addr, host.Type)
	}

	return err
}

func splitURL(URL string) (u *url.URL, auth proxy.Auth, scheme, host string) {
	var err error
	u, err = url.Parse(URL)
	if err != nil {
		return
	}
	scheme = u.Scheme
	host = u.Host
	if u.User != nil {
		auth.User = u.User.Username()
		auth.Password, _ = u.User.Password()
	}

	return u, auth, strings.ToLower(scheme), host
}
