package round

import (
	"fmt"
	"io/ioutil"
	"math/rand"
	"net"
	"net/http"
	"net/url"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/wrfly/gus-proxy/db"
	"github.com/wrfly/gus-proxy/prox"
	"github.com/wrfly/gus-proxy/types"
)

func TestRoundProxy(t *testing.T) {
	logrus.SetOutput(os.Stdout)

	ava := true
	hosts := []*types.ProxyHost{
		{
			Addr:      "socks5://127.0.0.1:1080",
			Available: ava,
		},
	}

	proxys, err := prox.New(hosts)
	assert.NoError(t, err)
	assert.NotNil(t, proxys)

	time.Sleep(1 * time.Second)
	l, err := net.Listen("tcp4", "127.0.0.1:8082")
	assert.NoError(t, err)
	assert.NotNil(t, l)

	DNSdb, err := db.New()
	if err != nil {
		logrus.Fatal(err)
	}
	defer DNSdb.Close()
	go http.Serve(l, New(proxys, DNSdb, ""))

	time.Sleep(6 * time.Second)
}

func TestCurlIPWithProxy(t *testing.T) {
	logrus.SetOutput(os.Stdout)

	ava := true
	localProxy := "127.0.0.1:8081"
	hosts := []*types.ProxyHost{
		{
			Addr:      "http://127.0.0.1:1081",
			Available: ava,
		},
		{
			Addr:      "socks5://127.0.0.1:1080",
			Available: ava,
		},
	}

	proxys, err := prox.New(hosts)
	assert.NoError(t, err)
	assert.NotNil(t, proxys)

	time.Sleep(1 * time.Second)
	l, err := net.Listen("tcp4", localProxy)
	assert.NoError(t, err)
	assert.NotNil(t, l)

	DNSdb, err := db.New()
	if err != nil {
		logrus.Fatal(err)
	}
	defer DNSdb.Close()
	go http.Serve(l, New(proxys, DNSdb, ""))

	proxyURL, _ := url.Parse("http://localhost:8081")
	clnt := &http.Client{
		Transport: &http.Transport{
			Proxy: http.ProxyURL(proxyURL),
		},
		Timeout: 3 * time.Second,
	}
	var wg sync.WaitGroup
	for i := 0; i < 6; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			resp, err := clnt.Get("http://ip.chinaz.com/getip.aspx")
			assert.NoError(t, err)
			if resp == nil {
				return
			}
			body, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				logrus.Error(err)
			}
			fmt.Printf("%s\n", body)
			resp.Body.Close()
		}()
	}
	wg.Wait()
}

func generateProxys(num int) []*types.ProxyHost {
	proxys := []*types.ProxyHost{}
	for i := 0; i < num; i++ {
		source := rand.NewSource(time.Now().UnixNano())
		r := rand.New(source)
		ping := r.Float32() * 10
		rProxy := &types.ProxyHost{
			Addr:      fmt.Sprintf("http://proxy-%d-ping-%f", i, ping),
			Ping:      ping,
			Available: true,
		}
		proxys = append(proxys, rProxy)
	}
	return proxys
}

func TestPingProxy(t *testing.T) {
	p := &Proxy{
		ProxyHosts: generateProxys(10),
		Scheduler:  PING,
	}
	rProxy := p.pingProxy()
	fmt.Println("Select: ", rProxy.Addr)

	fmt.Println("Proxy0: ", p.ProxyHosts[0].Addr)
}
