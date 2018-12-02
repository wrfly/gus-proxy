package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"sync"
	"sync/atomic"
	"time"
)

type IPinfoJson struct {
	IP       string `json:"ip,omitempty"`
	HOSTNAME string `json:"hostname,omitempty"`
	CITY     string `json:"city,omitempty"`
	REGION   string `json:"region,omitempty"`
	COUNTRY  string `json:"country,omitempty"`
	LOC      string `json:"loc,omitempty"`
	POSTAL   string `json:"postal,omitempty"`
	ORG      string `json:"org,omitempty"`
}

var logE = log.New(os.Stderr, "[ERR] ", log.Ltime)
var logI = log.New(os.Stderr, "[INFO] ", log.Ltime)

func main() {
	c := http.Client{
		Transport: &http.Transport{
			Proxy: http.ProxyFromEnvironment,
		},
		Timeout: time.Second * 5,
	}

	var (
		wg        sync.WaitGroup
		successed uint32
		failed    uint32
		ipMap     = make(map[string]bool, 100)

		start = time.Now()
	)

	for i := 0; i < 1e2; i++ {
		wg.Add(1)
		go func() {
			var (
				resp *http.Response
				x    = &IPinfoJson{}
				bs   []byte
				err  error
			)
			defer func() {
				wg.Done()
				if err != nil || x == nil {
					atomic.AddUint32(&failed, 1)
					if err == nil {
						logE.Printf("bad length of bytes: %d\n", len(bs))
					} else {
						logE.Println(err)
						logE.Printf("%s", bs)
					}
					return
				}
				remoteIP := x.IP
				logI.Printf("IP: %s", remoteIP)
				ipMap[remoteIP] = true
				atomic.AddUint32(&successed, 1)
			}()

			resp, err = c.Get("http://ipinfo.io")
			if err != nil {
				return
			}
			bs, err = ioutil.ReadAll(resp.Body)
			if err != nil {
				return
			}
			err = json.Unmarshal(bs, x)
			resp.Body.Close()
		}()
	}

	wg.Wait()
	logI.Printf("UniqIP: %d, OK: %d, Bad: %d, Use: %s",
		len(ipMap), successed, failed, use(start))
}

func use(s time.Time) time.Duration {
	return time.Now().Sub(s).Truncate(time.Millisecond)
}
