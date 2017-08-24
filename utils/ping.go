package utils

import (
	"net/url"
	"os/exec"
	"strconv"
	"strings"

	// ping "github.com/sparrc/go-ping"
	"github.com/sirupsen/logrus"
	"github.com/wrfly/gus-proxy/types"
)

// GetProxyPing ping the proxy ip and returns the average rtt
func GetProxyPing(host *types.ProxyHost) float32 {
	URL, _ := url.Parse(host.Addr)
	ip := strings.Split(URL.Host, ":")[0]
	logrus.Debugf("GetProxyPing [%s]", ip)

	cmd := exec.Command("ping", "-A", "-c", "3", "-w", "2", ip)
	b, err := cmd.Output()
	if err != nil {
		return 9999
	}

	o := string(b)
	l := strings.LastIndex(o, "=")
	o = string(o[l+2:])
	s := strings.Split(o, "/")

	avg := s[1]

	f, _ := strconv.ParseFloat(avg, 32)

	return float32(f)
}
