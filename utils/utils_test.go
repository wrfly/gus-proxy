package utils

import (
	"fmt"
	"strings"
	"testing"
)

func TestPing(t *testing.T) {
	t.Log(Ping("8.8.8.8"))
	t.Log(Ping("1.1.1.1"))
	t.Log(Ping("kfd.me"))
}

func TestParsePing(t *testing.T) {
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

func TestSelectUA(t *testing.T) {
	ua1, ua2 := RandomUA(), RandomUA()
	if ua1 == ua2 {
		t.Error("random ua failed")
	}
}

func TestHashSlice(t *testing.T) {
	slice := []string{"1", "2", "3"}
	if HashSlice(slice) != HashSlice(slice) {
		t.Error("not equal")
	}

	if HashSlice(slice) == HashSlice(append(slice, "4")) {
		t.Error("equal")
	}
}

func TestPubIP(t *testing.T) {
	ip, err := PublicIP()
	if err != nil {
		t.Error(err)
	}
	t.Log(ip)
}
