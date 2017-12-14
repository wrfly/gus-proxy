# Gus-Proxy

"gus - 绝命毒师里的大毒枭"

"gus - the heavy-duty drug trafficker in *Breaking Bad*"

[![Build Status](https://travis-ci.org/wrfly/gus-proxy.svg?branch=master)](https://travis-ci.org/wrfly/gus-proxy)

---

## Thoughts

> 打一枪换一个地方

1. 每次请求都从代理池中选取一个代理
1. 但是这样会不会触发server端的验证，即session与IP匹配
1. 但是如果server端有这种IP验证的话，就没必要用这东西了
1. 要解决的是server限制某一IP访问频率的问题

没问题。

> Change our IP address every request

1. Chose a different proxy in our proxy poll every request
1. If our IP changed, the server side may not auth us because of the session-IP pair
1. No use for session authentication
1. The aim for this tool is to resolve the restrict of IP request limit

Ok.

## Design

1. 程序对上层表现为一个HTTP代理
1. 程序加载一个代理列表（HTTP/Socks5） [或者默认配置一个代理列表]
1. 每次的请求都从代理列表中选取一个
1. 选取的算法可能是轮询、随机、或其他目前没想到的
1. 要验证proxy的可用性
1. 每次请求替换UA
1. 请求资源的时候，查询目标资源地址全部的IP，随机

1. An top layer HTTP-proxy
1. The program load a proxy list(HTTP or Socks5) during start
1. Chose a proxy every request
1. May have different choose algorithm: round-robin|random|ping
1. Verify the availability of the proxy
1. Change our UA every request(it's an option)
1. Lookup target's all IP address, replace target host at random every request 

## Show off

![Gus-Running](img/gus-run.png)
![Curl-test](img/gus-curl.png)

## Run

```bash
sudo docker run --rm -ti -p 8080:8080 wrfly/gus-proxy
```