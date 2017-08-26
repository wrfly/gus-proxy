test:
	go test -v `glide nv`

build:
	go build -o gus-proxy

local:
	rm -rf vendor/github.com/wrfly/goproxy
	go test -v `glide nv`
	go build -o gus-proxy
	./scripts/kuaidaili.sh > proxyhosts_test.txt
	./gus-proxy -f proxyhosts_test.txt
curl:
	for i in 1 2 3 4 5 ;do \
		curl --proxy http://localhost:8080 ip.chinaz.com/getip.aspx ; \
	done