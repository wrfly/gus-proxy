package round

import (
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/wrfly/gus-proxy/prox"
	"github.com/wrfly/gus-proxy/types"
)

func TestRoundProxy(t *testing.T) {
	log.SetOutput(os.Stdout)

	hosts := []*types.ProxyHost{
		&types.ProxyHost{Addr: "http://61.130.97.212:8099"},
		&types.ProxyHost{Addr: "socks5://127.0.0.1:1080"},
	}

	proxys, err := prox.New(hosts)
	assert.NoError(t, err)
	assert.NotNil(t, proxys)

	logrus.Info(proxys)

	l, err := net.Listen("tcp4", "127.0.0.1:8080")
	assert.NoError(t, err)
	go http.Serve(l, New(proxys))

	time.Sleep(10 * time.Second)
}

func TestCurlIPWithProxy(t *testing.T) {
	localProxy := "127.0.0.1:8080"
	hosts := []*types.ProxyHost{
		&types.ProxyHost{Addr: "http://61.130.97.212:8099"},
		&types.ProxyHost{Addr: "socks5://127.0.0.1:1080"},
	}

	proxys, err := prox.New(hosts)
	assert.NoError(t, err)
	assert.NotNil(t, proxys)

	logrus.Info(proxys)

	l, err := net.Listen("tcp4", localProxy)
	assert.NoError(t, err)
	go http.Serve(l, New(proxys))

	proxyURL, _ := url.Parse("http://localhost:8080")
	clnt := &http.Client{
		Transport: &http.Transport{
			Proxy: http.ProxyURL(proxyURL),
		},
	}
	for i := 0; i < 9; i++ {
		resp, err := clnt.Get("http://ip.chinaz.com/getip.aspx")
		assert.NoError(t, err)
		body, _ := ioutil.ReadAll(resp.Body)
		fmt.Printf("%s\n", body)
	}
}
