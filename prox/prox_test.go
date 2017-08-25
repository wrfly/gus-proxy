package prox

import (
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/wrfly/gus-proxy/types"
	"golang.org/x/net/proxy"
)

func TestHttpProxy(t *testing.T) {
	addr := "http://127.0.0.1:1081"
	p, err := proxyHTTP(addr)
	assert.NoError(t, err)

	go http.ListenAndServe("127.0.0.1:8080", p)
	time.Sleep(10 * time.Second)
}

func TestSocks5Proxy(t *testing.T) {
	addr := "127.0.0.1:1080"
	p, err := proxySocks5(addr, proxy.Auth{})
	assert.NoError(t, err)

	go http.ListenAndServe("127.0.0.1:8080", p)
	time.Sleep(10 * time.Second)
}

func TestNew(t *testing.T) {
	logrus.SetLevel(logrus.DebugLevel)
	logrus.SetOutput(os.Stdout)

	hosts := []*types.ProxyHost{
		&types.ProxyHost{Addr: "socks5://127.0.0.1:1080"},
		&types.ProxyHost{Addr: "https://127.0.0.1:1081"},
	}
	proxs, err := New(hosts)
	assert.NoError(t, err)
	assert.NotNil(t, proxs)
	t.Log(proxs)
}

func TestSplitURL(t *testing.T) {
	URLS := []string{
		"http://localhost:8080",
		"http://u:p@localhost:8080",
		"socks5://localhost:8080",
		"socks5://usner:ocxasd@localhost:8080",
		"https://usnerocxasd@localhost:8080",
	}
	for _, U := range URLS {
		auth, scheme, hostAndPort := splitURL(U)
		t.Log(auth, scheme, hostAndPort)
	}
}
