# go-socks4
Socks4 implementation for Go, compatible with net/proxy

## Usage
```go
import (
	"golang.org/x/net/proxy"
	_ "github.com/Bogdan-D/go-socks4"
)

func main() {
	dialer, err := proxy.FromURL("socks4://ip:port",proxy.Direct)
	// check error
	// and use your dialer as you with
}
```


## Tests
If you know proxy server to connect to tests should be running like this
`
go test -socks4.address=localhost:8080
`




