test:
	go test -cover -v `glide nv`

build:
	go build -o gus-proxy

dev:
	go build -o gus-proxy
	./gus-proxy -f proxyhosts_test.txt -d

curl:
	for i in 1 2 3 4 5 ;do \
		curl --proxy http://localhost:8080 ip.chinaz.com/getip.aspx ; \
	done