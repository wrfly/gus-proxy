# Gus-Proxy

"gus - the heavy-duty drug trafficker in *Breaking Bad*"

[![Build Status](https://travis-ci.org/wrfly/gus-proxy.svg?branch=master)](https://travis-ci.org/wrfly/gus-proxy)
[![Go Report Card](https://goreportcard.com/badge/github.com/wrfly/gus-proxy)](https://goreportcard.com/report/github.com/wrfly/gus-proxy)

[README.Chinese](README.md)

---

## Run

```bash
# prepare the proxies
mkdir -p data
touch data/proxies.txt
# put your proxies in data/proxies.txt
# the format could be:
# socks5://127.0.0.1:1080
# http://user:pass@127.0.0.1:1081
# socks4://127.0.0.1:1082
# direct://0.0.0.0

# then:
sudo docker run --rm -ti --name gus-proxy \
    -p 8080:8080 \
    -v `pwd`/data:/data \
    wrfly/gus-proxy
```

## Thoughts

> Change our IP address every request

1. Chose a different proxy in our proxy poll every request
1. If our IP changed, the server side may not auth us because of the session-IP pair
1. No use for session authentication
1. The aim for this tool is to resolve the restrict of IP request limit

## Design

1. An top layer HTTP-proxy
1. The program load a proxy list(HTTP or Socks5) during start
1. Chose a proxy every request
1. May have different choose algorithm: round-robin|random|ping
1. Verify the availability of the proxy
1. Change our UA every request(it's an option)
1. Lookup target's all IP address, replace target host per request

## Show off

![Gus-Running](img/gus-run.png)
![Curl-test](img/gus-curl.png)
