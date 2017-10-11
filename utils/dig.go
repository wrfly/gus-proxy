package utils

import (
	"fmt"
	"math/rand"
	"net"
	"strings"
	"time"

	"github.com/miekg/dns"
	"github.com/sirupsen/logrus"
	"github.com/wrfly/gus-proxy/db"
)

// SelectIP kfd.me -> x.x.x.x
// hello.cn:8080 -> xx.x.xx.xx:8080
func SelectIP(host string, dnsDB *db.DNS) string {
	str := strings.Split(host, ":")
	domain := str[0]
	port := "80"
	if len(str) == 2 {
		port = str[1]
	}

	ips := dnsDB.Query(domain)
	logrus.Debugf("Query %s IP: %v", domain, ips)
	// not found in db
	if len(ips) == 0 {
		digIPs, err := dig(domain) // shadow value
		if err != nil {
			logrus.Errorf("Dig Error: %s", err)
			return "127.0.0.1"
		}
		// set to db
		logrus.Debugf("Set DNS DB: domain: %s IP: %v", domain, digIPs)
		if err := dnsDB.SetDNS(domain, digIPs); err != nil {
			logrus.Error(err)
		}
		ips = digIPs
	}

	i := rand.Int() % len(ips)
	ip := ips[i]
	ip = fmt.Sprintf("%s:%s", ip, port)

	return ip
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
