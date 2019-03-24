# Gus-Proxy

"gus - 绝命毒师里的大毒枭"

[![Build Status](https://travis-ci.org/wrfly/gus-proxy.svg?branch=master)](https://travis-ci.org/wrfly/gus-proxy)
[![Go Report Card](https://goreportcard.com/badge/github.com/wrfly/gus-proxy)](https://goreportcard.com/report/github.com/wrfly/gus-proxy)

[README.English](README.en.md)

---

## 运行

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

## 思路

> 打一枪换一个地方

1. 每次请求都从代理池中选取一个代理
1. 但是这样会不会触发server端的验证，即session与IP匹配
1. 但是如果server端有这种IP验证的话，就没必要用这东西了
1. 要解决的是server限制某一IP访问频率的问题

没问题。

## 设计

1. 程序对上层表现为一个HTTP代理
1. 程序加载一个代理列表（HTTP/Socks5） [或者默认配置一个代理列表]
1. 每次的请求都从代理列表中选取一个
1. 选取的算法可能是轮询、随机、或其他目前没想到的
1. 要验证proxy的可用性
1. 每次请求替换UA
1. 请求资源的时候，查询目标资源地址全部的IP，随机

## 效果

![Gus-Running](img/gus-run.png)
![Curl-test](img/gus-curl.png)
