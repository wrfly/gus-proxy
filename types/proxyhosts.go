package types

import (
	"github.com/wrfly/goproxy"
	"golang.org/x/net/proxy"
)

// ProxyHost defines the proxy
type ProxyHost struct {
	Type    string // http or socks5
	Addr    string // 127.0.0.1:1080
	Alive   bool
	Ping    float32 // 66 ms
	Auth    proxy.Auth
	GoProxy *goproxy.ProxyHttpServer
}
