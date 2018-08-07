NAME = gus-proxy
PKG = github.com/wrfly/$(NAME)
BIN = bin
IMAGE := wrfly/$(NAME)

VERSION := $(shell cat VERSION)
COMMITID := $(shell git rev-parse --short HEAD)
BUILDAT := $(shell date +%Y-%m-%d)

CTIMEVAR = -X main.CommitID=$(COMMITID) \
	-X main.Version=$(VERSION) \
	-X main.BuildAt=$(BUILDAT)
GO_LDFLAGS = -ldflags "-s -w $(CTIMEVAR)"

GLIDE_URL := https://github.com/Masterminds/glide/releases/download/v0.13.1/glide-v0.13.1-linux-amd64.tar.gz

.PHONY: prepare
prepare:
	wget $(GLIDE_URL) -O glide.tgz
	tar xzf glide.tgz
	linux-amd64/glide install

.PHONY: bin
bin:
	mkdir -p bin

.PHONY: glide-up
glide-up:
	https_proxy=http://127.0.0.1:1081 glide up

.PHONY: build
build: bin
	go build $(GO_LDFLAGS) -o $(BIN)/$(NAME) .

.PHONY: test
test:
	go test -cover -v `glide nv`

.PHONY: dev
dev:
	go build -o $(NAME)
	./$(NAME) run -f proxyhosts_test.txt -d

.PHONY: curl
curl:
	for i in 1 2 3 4 5 ;do \
		curl --proxy http://localhost:8080 ip.chinaz.com/getip.aspx ; \
	done

.PHONY: release
release:
	GOOS=linux GOARCH=amd64 go build $(GO_LDFLAGS) -o $(BIN)/$(NAME)_linux_amd64 .
	GOOS=darwin GOARCH=amd64 go build $(GO_LDFLAGS) -o $(BIN)/$(NAME)_darwin_amd64 .
	tar -C $(BIN) -czf $(BIN)/$(NAME)_linux_amd64.tgz $(NAME)_linux_amd64
	tar -C $(BIN) -czf $(BIN)/$(NAME)_darwin_amd64.tgz $(NAME)_darwin_amd64

.PHONY: image
image:
	docker build -t $(IMAGE) .

.PHONY: push-image
push-image:
	docker push $(IMAGE)

.PHONY: push-develop
push-develop:
	docker tag $(IMAGE) $(IMAGE):develop
	docker push $(IMAGE):develop

.PHONY: push-tag
push-tag:
	docker tag $(IMAGE) $(IMAGE):$(VERSION)
	docker push $(IMAGE):$(VERSION)
