package utils

import (
	"net"
	"net/http"
	"net/url"
	"time"

	"github.com/wrfly/gus-proxy/types"
)

// CheckProxyAvailable checks if the proxy is available
func CheckProxyAvailable(host types.ProxyHost) error {
	proxyURL, err := url.Parse(host.Addr)
	if err != nil {
		return err
	}

	conn, err := net.DialTimeout("tcp", proxyURL.Host, 1*time.Second)
	if err != nil {
		return err
	}
	defer conn.Close()

	clnt := &http.Client{
		Transport: &http.Transport{
			Proxy: http.ProxyURL(proxyURL),
		},
		Timeout: 3 * time.Second,
	}

	reqURL := &url.URL{
		Path:   "/getip.aspx",
		Scheme: "http",
		Host:   "112.5.254.80",
	}
	req := &http.Request{
		Method: "GET",
		URL:    reqURL,
		Host:   "ip.chinaz.com",
	}

	resp, err := clnt.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return nil
}
