package utils

import (
	"fmt"
	"math/rand"
	"net"
	"strings"

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
		digIPs, err := lookupHost(domain) // shadow value
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

	ip := ips[rand.Int()%len(ips)]
	return fmt.Sprintf("%s:%s", ip, port)
}

func lookupHost(domain string) (IPs []string, err error) {
	ips, err := net.LookupIP(domain)
	if err != nil {
		return nil, err
	}

	ipv4 := make([]string, 0, len(ips))
	for _, ip := range ips {
		if ip.To4() == nil { // not an ipv4 address
			continue
		}
		ipv4 = append(ipv4, ip.String())
	}

	return ipv4[:len(ipv4)], nil

}
