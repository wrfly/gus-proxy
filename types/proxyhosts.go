package types

import "github.com/wrfly/goproxy"

// ProxyHost defines the proxy
type ProxyHost struct {
	Type     string // http or socks5
	Addr     string // 127.0.0.1:1080
	Weight   int
	LastPing int
	Alive    bool
	Auth     ProxyAuth
	GoProxy  *goproxy.ProxyHttpServer
}

// ProxyAuth username@passwd
type ProxyAuth struct {
	Username string
	Password string
}
