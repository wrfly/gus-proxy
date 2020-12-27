package proxy

// supported chose method
const (
	RR     = "round_robin"
	RANDOM = "random"
	PING   = "ping"
)

// supported protocols
const (
	DIRECT = "direct"
	HTTP   = "http"
	HTTPS  = "https"
	SOCKS4 = "socks4"
	SOCKS5 = "socks5"

	// shadowsocks protocol
	ShadorSocks = "ss"
)
