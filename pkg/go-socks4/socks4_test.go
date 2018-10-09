package socks4

import (
	"bufio"
	"flag"
	"fmt"
	"golang.org/x/net/proxy"
	"net/url"
	"testing"
)

var address string

func init() {
	flag.StringVar(&address, "socks4.address", "", "address of socks4 server to connect to")
	flag.Parse()
}

func TestDial(t *testing.T) {
	proxy_addr, _ := url.Parse(address)

	var socks = &socks4{url: proxy_addr, dialer: proxy.Direct}

	c, err := socks.Dial("tcp", "google.com:80")

	if err != nil {
		e, _ := err.(*socks4Error)

		switch e.String() {
		case ErrIdentRequired:
		default:
			t.Error(err)
		}
	} else {
		_, err := c.Write([]byte("GET /\n"))
		if err != nil {
			t.Error(err)
		}
		buf := bufio.NewReader(c)
		line, err := buf.ReadString('\n')
		fmt.Print(line)
		c.Close()
	}
}
