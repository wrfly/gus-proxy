package types

// Host defines the proxy host
type Host struct {
	Type   string // http or socks5
	Addr   string // 127.0.0.1:1080
	Weight int
	Alive  bool
}
