package utils

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/wrfly/gus-proxy/types"
)

func TestPing(t *testing.T) {
	host := &types.ProxyHost{
		Addr: "http://8.8.8.8:1080",
	}
	t.Log(GetProxyPing(host))
}

func Test(t *testing.T) {
	o := `PING 8.8.8.8 (8.8.8.8) 56(84) bytes of data.
64 bytes from 8.8.8.8: icmp_seq=1 ttl=39 time=63.4 ms
64 bytes from 8.8.8.8: icmp_seq=2 ttl=39 time=59.4 ms
64 bytes from 8.8.8.8: icmp_seq=3 ttl=39 time=69.9 ms

--- 8.8.8.8 ping statistics ---
3 packets transmitted, 3 received, 0% packet loss, time 401ms
rtt min/avg/max/mdev = 59.498/64.273/69.909/4.303 ms, ipg/ewma 200.829/63.797 ms`

	l := strings.LastIndex(o, "=")
	o = string(o[l+2:])
	s := strings.Split(o, "/")

	fmt.Println(s[1])

}

func TestCheckProxyAvailable(t *testing.T) {
	host := types.ProxyHost{
		// Addr: "http://61.130.97.212:8099", // www.baidu.com
		Addr: "http://119.75.216.20:80", // www.baidu.com
	}
	err := CheckProxyAvailable(&host)
	assert.Error(t, err)
}

func TestDig(t *testing.T) {
	t.Log(SelectIP("kfd.me"))
	t.Log(SelectIP("kfd.me:8080"))
	t.Log(SelectIP("kfd.me"))
}

func TestRandomUA(t *testing.T) {
	t.Log(RandomUA())
}
