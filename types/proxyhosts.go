package types

import (
	"github.com/elazarl/goproxy"
	"golang.org/x/net/proxy"
)

// ProxyHost defines the proxy
type ProxyHost struct {
	Type      string  // http or socks5 or direct
	Addr      string  // 127.0.0.1:1080
	Ping      float32 // 66 ms
	Available bool
	Auth      proxy.Auth
	GoProxy   *goproxy.ProxyHttpServer
}
