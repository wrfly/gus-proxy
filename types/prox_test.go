package types

import (
	"io/ioutil"
	"net/http"
	"os"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"golang.org/x/net/proxy"
)

func TestHttpProxy(t *testing.T) {
	addr := "http://127.0.0.1:1081"
	p, err := proxyHTTP(addr)
	assert.NoError(t, err)

	c := http.Client{Transport: p.Tr}
	r, _ := http.NewRequest("GET", "http://ipinfo.io", nil)
	r.Header.Set("User-Agent", "curl")
	resp, err := c.Do(r)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	bs, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("%s", bs)
}

func TestSocks5Proxy(t *testing.T) {
	addr := "127.0.0.1:1080"
	p, err := proxySocks5(addr, proxy.Auth{})
	assert.NoError(t, err)

	c := http.Client{Transport: p.Tr}
	r, _ := http.NewRequest("GET", "http://ipinfo.io", nil)
	r.Header.Set("User-Agent", "curl")
	resp, err := c.Do(r)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	bs, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("%s", bs)
}

func TestDirect(t *testing.T) {
	p := proxyDirect()

	c := http.Client{Transport: p.Tr}
	r, _ := http.NewRequest("GET", "http://ipinfo.io", nil)
	r.Header.Set("User-Agent", "curl")
	resp, err := c.Do(r)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	bs, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("%s", bs)

}

func TestInitProxy(t *testing.T) {
	logrus.SetLevel(logrus.DebugLevel)
	logrus.SetOutput(os.Stdout)

	hosts := []*ProxyHost{
		{Addr: "socks5://127.0.0.1:1080"},
		{Addr: "https://127.0.0.1:1081"},
		{Addr: "direct://0.0.0.0"},
	}
	for _, host := range hosts {
		err := host.Init()
		assert.NoError(t, err)
		t.Logf("%s: %v\n", host.Addr, host.Available)
	}
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
