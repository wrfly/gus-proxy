.PHONY: test build dev curl

NAME := "gus-proxy"

test:
	go test -cover -v `glide nv`

build:
	go build -o $(NAME)

dev:
	go build -o $(NAME)
	./$(NAME) run -f proxyhosts_test.txt -d

curl:
	for i in 1 2 3 4 5 ;do \
		curl --proxy http://localhost:8080 ip.chinaz.com/getip.aspx ; \
	done