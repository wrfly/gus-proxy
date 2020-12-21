package gus

import (
	"fmt"
	"io/ioutil"
	"math/rand"
	"net"
	"net/http"
	"net/url"
	"sync"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/wrfly/gus-proxy/config"
	"github.com/wrfly/gus-proxy/db"
	"github.com/wrfly/gus-proxy/proxy"
)

func TestRoundProxy(t *testing.T) {
	logrus.SetLevel(logrus.DebugLevel)

	testNewProxy := func(t *testing.T) {
		ava := true
		hosts := []*proxy.Host{
			{
				Addr:      "socks5://127.0.0.1:1080",
				Available: ava,
			},
		}
		for _, host := range hosts {
			err := host.Init()
			assert.NoError(t, err)
		}

	}
	t.Run("test new proxy", testNewProxy)

	time.Sleep(1 * time.Second) // release port binding

	serveProxy := func(t *testing.T) {
		// mock config
		c := &config.Config{
			ListenPort:    "54321",
			ProxyFilePath: "../proxyhosts.txt",
			Scheduler:     proxy.ROUND_ROBIN,
			RandomUA:      true,
		}
		err := c.Validate()
		assert.NoError(t, err)
		c.UpdateProxies()

		l, err := net.Listen("tcp4", fmt.Sprintf(":%s", c.ListenPort))
		assert.NoError(t, err)
		assert.NotNil(t, l)

		DNSdb, err := db.New("/tmp/gus.db")
		if err != nil {
			logrus.Fatalf("init dns db error: %s", err)
		}
		defer DNSdb.Close()

		go http.Serve(l, New(c, DNSdb))
	}
	t.Run("serve proxy", serveProxy)

	time.Sleep(time.Second)
	testWithCurl := func(t *testing.T) {
		proxyURL, _ := url.Parse("http://localhost:54321")
		clnt := &http.Client{
			Transport: &http.Transport{
				Proxy: http.ProxyURL(proxyURL),
			},
			Timeout: 5 * time.Second,
		}
		var wg sync.WaitGroup
		for i := 0; i < 6; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				resp, err := clnt.Get("http://ip.kfd.me")
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

	t.Run("with curl", testWithCurl)
}

func generateProxys() []*proxy.Host {
	num := 10
	proxys := []*proxy.Host{}
	for i := 0; i < num; i++ {
		source := rand.NewSource(time.Now().UnixNano())
		r := rand.New(source)
		ping := r.Float32() * 10
		rProxy := &proxy.Host{
			Addr:      fmt.Sprintf("http://proxy-%d-ping-%f", i, ping),
			Ping:      ping,
			Available: true,
		}
		proxys = append(proxys, rProxy)
	}
	return proxys
}

func TestPingProxy(t *testing.T) {
	p := &Gustavo{
		proxyHosts: generateProxys,
		scheduler:  proxy.PING,
	}
	rProxy := p.pingProxy()
	fmt.Println("Select: ", rProxy.Addr)

	fmt.Println("Proxy0: ", p.proxyHosts()[0].Addr)
}
