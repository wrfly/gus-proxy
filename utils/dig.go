package utils

import (
	"net"
)

func LookupHost(domain string) (IPs []string, err error) {
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
