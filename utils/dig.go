package utils

import (
	"fmt"
	"math/rand"
	"net"
	"strings"
	"time"

	"github.com/miekg/dns"
	"github.com/sirupsen/logrus"
)

// SelectIP kfd.me -> x.x.x.x
// hello.cn:8080 -> xx.x.xx.xx:8080
func SelectIP(host string) string {
	s := strings.Split(host, ":")
	ips, err := dig(s[0])
	if err != nil {
		logrus.Errorf("Dig Error: %s", err)
		return "127.0.0.1"
	}

	i := rand.Int() % len(ips)
	if strings.Contains(host, ":") {
		return fmt.Sprintf("%s:%s", ips[i], s[1])
	}
	return ips[i]
}

func dig(domain string) (IPs []string, err error) {
	config, _ := dns.ClientConfigFromFile("/etc/resolv.conf")
	c := new(dns.Client)
	c.DialTimeout = time.Duration(2 * time.Second)
	c.ReadTimeout = time.Duration(2 * time.Second)

	m := new(dns.Msg)
	m.SetQuestion(dns.Fqdn(domain), dns.TypeA)
	m.RecursionDesired = true

	r, _, err := c.Exchange(m, net.JoinHostPort(config.Servers[0], config.Port))
	if r == nil || err != nil {
		return nil, fmt.Errorf("answer is nil; err: %s", err)
	}

	if r.Rcode != dns.RcodeSuccess {
		return nil, fmt.Errorf("dig domain [%s] failed", domain)
	}

	for _, a := range r.Answer {
		ip := getIPOnly(a.String())
		IPs = append(IPs, ip)
	}
	return IPs, nil
}

func getIPOnly(answer string) (ip string) {
	tabPos := strings.LastIndex(answer, "\t")
	return string(answer[tabPos+1:])
}
