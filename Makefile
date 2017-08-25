test:
	go test `glide nv`

build:
	go build -o gus-proxy

local:
	go build -o gus-proxy
	./gus-proxy -f proxyhosts_test.txt