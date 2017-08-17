package config

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"net/url"
	"os"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/wrfly/gus-proxy/round"
	"github.com/wrfly/gus-proxy/types"
)

// Config ...
type Config struct {
	ProxyHostsFile string
	Scheduler      string
	ListenPort     string
}

// Validate the config
func (c *Config) Validate() bool {
	_, err := os.Open(c.ProxyHostsFile)
	if err != nil && os.IsNotExist(err) {
		logrus.Errorf("Hostfile [%s] not exist", c.ProxyHostsFile)
		return false
	}
	switch c.Scheduler {
	case round.ROUND_ROBIN:
	case round.RANDOM:
	case round.PING:
	default:
		return false
	}

	// listen port
	c.ListenPort = fmt.Sprintf(":%s", c.ListenPort)
	l, err := net.Listen("tcp4", c.ListenPort)
	if err != nil {
		logrus.Errorf("Can not bind this port: %s", err)
		return false
	}
	defer l.Close()
	return true
}

// LoadHosts returns the proxy hosts
func (c *Config) LoadHosts() ([]*types.ProxyHost, error) {
	proxyHosts := []*types.ProxyHost{}
	f, _ := os.Open(c.ProxyHostsFile)
	r := bufio.NewReader(f)
	for {
		s, err := r.ReadString('\n')
		if err != nil && s == "" {
			if err == io.EOF {
				break
			}
			return nil, err
		}

		// verify hosts
		s = strings.TrimRight(s, "\n")
		_, err = url.Parse(s)
		if err != nil {
			return nil, err
		}

		host := &types.ProxyHost{
			Addr: s,
		}
		proxyHosts = append(proxyHosts, host)
	}

	return proxyHosts, nil
}
