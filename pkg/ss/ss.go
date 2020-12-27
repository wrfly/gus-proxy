package ss

import (
	"fmt"
	"net"
	"net/url"

	"github.com/shadowsocks/go-shadowsocks2/core"
	"github.com/shadowsocks/go-shadowsocks2/socks"
	"github.com/sirupsen/logrus"
	"golang.org/x/net/proxy"
)

func init() {
	proxy.RegisterDialerType("ss", ssDialer)
}

func ssDialer(u *url.URL, d proxy.Dialer) (proxy.Dialer, error) {
	if u.User == nil {
		return nil, fmt.Errorf("empty user")
	}

	ss := new(shadowsocks)
	ss.dialer = d

	cipherName := u.User.Username()
	ss.server = u.Host
	passwd, _ := u.User.Password()

	logrus.Debugf("init url: %+v %s", u, cipherName)
	cipher, err := core.PickCipher(cipherName, []byte{}, passwd)
	if err != nil {
		return nil, fmt.Errorf("pick cipher err: %s", err)
	}
	ss.cipher = cipher

	return ss, nil
}

type shadowsocks struct {
	dialer proxy.Dialer

	cipher core.Cipher
	server string
}

func (ss *shadowsocks) Dial(network, addr string) (net.Conn, error) {
	if network != "tcp" {
		return ss.dialer.Dial(network, addr)
	}

	rc, err := net.Dial("tcp", ss.server)
	if err != nil {
		return nil, fmt.Errorf("dial target err: %s", err)
	}
	rc = ss.cipher.StreamConn(rc)
	rc.Write(socks.ParseAddr(addr))
	return rc, nil
}
