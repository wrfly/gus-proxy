package socks4

import (
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"net/url"
	"strconv"

	"golang.org/x/net/proxy"
)

const (
	socks_version = 0x04
	socks_connect = 0x01
	socks_bind    = 0x02

	socks_ident = "nobody@0.0.0.0"

	access_granted         = 0x5a
	access_rejected        = 0x5b
	access_identd_required = 0x5c
	access_identd_failed   = 0x5d

	ErrWrongURL      = "wrong server url"
	ErrWrongConnType = "no support for connections of type"
	ErrConnFailed    = "connection failed to socks4 server"
	ErrHostUnknown   = "unable to find IP address of host"
	ErrSocksServer   = "socks4 server error"
	ErrConnRejected  = "connection rejected"
	ErrIdentRequired = "socks4 server require valid identd"
)

func init() {
	proxy.RegisterDialerType("socks4", func(u *url.URL, d proxy.Dialer) (proxy.Dialer, error) {
		return &socks4{url: u, dialer: d}, nil
	})
}

type socks4Error struct {
	message string
	details interface{}
}

func (s *socks4Error) String() string {
	return s.message
}

func (s *socks4Error) Error() string {
	if s.details == nil {
		return s.message
	}

	return fmt.Sprintf("%s: %v", s.message, s.details)
}

type socks4 struct {
	url    *url.URL
	dialer proxy.Dialer
}

func (s *socks4) Dial(network, addr string) (c net.Conn, err error) {
	var buf []byte

	switch network {
	case "tcp", "tcp4":
	default:
		return nil, &socks4Error{message: ErrWrongConnType, details: network}
	}

	c, err = s.dialer.Dial(network, s.url.Host)
	if err != nil {
		return nil, &socks4Error{message: ErrConnFailed, details: err}
	}

	host, port, err := net.SplitHostPort(addr)
	if err != nil {
		return nil, &socks4Error{message: ErrWrongURL, details: err}
	}

	ip, err := net.ResolveIPAddr("ip4", host)
	if err != nil {
		return nil, &socks4Error{message: ErrHostUnknown, details: err}
	}
	ip4 := ip.IP.To4()

	var bport [2]byte
	iport, _ := strconv.Atoi(port)
	binary.BigEndian.PutUint16(bport[:], uint16(iport))

	buf = []byte{socks_version, socks_connect}
	buf = append(buf, bport[:]...)
	buf = append(buf, ip4...)
	buf = append(buf, socks_ident...)
	buf = append(buf, 0)

	i, err := c.Write(buf)
	if err != nil {
		return nil, &socks4Error{message: ErrSocksServer, details: err}
	}
	if l := len(buf); i != l {
		return nil, &socks4Error{message: ErrSocksServer, details: fmt.Sprintf("write %d bytes, expected %d", i, l)}
	}

	var resp [8]byte
	i, err = c.Read(resp[:])
	if err != nil && err != io.EOF {
		return nil, &socks4Error{message: ErrSocksServer, details: err}
	}
	if i != 8 {
		return nil, &socks4Error{message: ErrSocksServer, details: fmt.Sprintf("read %d bytes, expected 8", i)}
	}

	switch resp[1] {
	case access_granted:
		return c, nil
	case access_identd_required, access_identd_failed:
		return nil, &socks4Error{message: ErrIdentRequired, details: strconv.FormatInt(int64(resp[1]), 16)}
	default:
		c.Close()
		return nil, &socks4Error{message: ErrConnRejected, details: strconv.FormatInt(int64(resp[1]), 16)}
	}
}
