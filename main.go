package main

import (
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/elazarl/goproxy"
	log "github.com/sirupsen/logrus"
	"golang.org/x/net/proxy"
)

// func main() {

// 	proxy, err := proxySocks5("127.0.0.1:1080")
// 	if err != nil {
// 		log.Errorf("make proxy error: [%s]", err)
// 	}
// 	proxy.
// 		log.Fatal(http.ListenAndServe(":8080", nil))
// }

func proxyHTTP(httpAddr string) (*goproxy.ProxyHttpServer, error) {
	// goproxy.NewProxyHttpServer()
	return nil, nil
}

func proxySocks5(socks5Addr string) (*goproxy.ProxyHttpServer, error) {
	dialer, err := proxy.SOCKS5("tcp", socks5Addr, nil, proxy.Direct)
	if err != nil {
		return nil, err
	}

	proxy := goproxy.NewProxyHttpServer()
	proxy.Verbose = true
	proxy.Tr.Dial = dialer.Dial
	return proxy, nil
}

func main() {
	hf := http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		c := &http.Client{}

		// reset header
		req.Header.Set("User-Agent", "google bot")
		req.RequestURI = ""

		p := "http://"
		if req.TLS != nil {
			p = "https://"
		}
		h := req.Host
		i := req.URL.Path

		u, err := url.Parse(fmt.Sprintf("%s%s%s", p, h, i))
		if err != nil {
			log.Error(err)
			return
		}
		req.URL = u

		resp, err := c.Do(req)
		if err != nil {
			log.Error(err)
			return
		}
		copyHeaders(w.Header(), resp.Header)
		w.WriteHeader(resp.StatusCode)
		_, err = io.Copy(w, resp.Body)
		if err != nil {
			w.WriteHeader(500)
			w.Write([]byte(err.Error()))
			return
		}
		resp.Body.Close()
	})

	http.HandleFunc("/", hf)
	http.ListenAndServe(":8080", nil)
}

func copyHeaders(dst, src http.Header) {
	for k := range dst {
		dst.Del(k)
	}
	for k, vs := range src {
		for _, v := range vs {
			dst.Add(k, v)
		}
	}
}

func resetRequest(ori *http.Request) *http.Request {
	r := &http.Request{
		Method:     ori.Method,
		URL:        ori.URL,
		Proto:      ori.Proto,
		ProtoMajor: ori.ProtoMajor,
		ProtoMinor: ori.ProtoMinor,
		Header:     ori.Header,
	}
	r.Header.Set("User-Agent", "google bot")

	return r
}
