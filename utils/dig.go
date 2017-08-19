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

// SelectIP select one ip form dig answers
func SelectIP(domain string) string {
	ips, err := DigIP(domain)
	if err != nil {
		logrus.Errorf("Select IP Error: %s", err)
		return "127.0.0.1"
	}

	i := rand.Int() % len(ips)
	return ips[i]
}

// DigIP domain's IPs
func DigIP(domain string) (IPs []string, err error) {
	config, _ := dns.ClientConfigFromFile("/etc/resolv.conf")
	c := new(dns.Client)
	c.DialTimeout = time.Duration(3 * time.Second)
	c.ReadTimeout = time.Duration(3 * time.Second)

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
