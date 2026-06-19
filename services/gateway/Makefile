.PHONY: all build build-linux test test-short vet lint tidy clean proto wire docker-build docker-push help

GO              ?= go
PROTOC          ?= protoc
WIRE            ?= wire
VERSION         ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
IMAGE_VERSION   ?= $(VERSION)
ACR_REGISTRY    ?= halei-acr-new-registry.eu-central-1.cr.aliyuncs.com
ACR_NAMESPACE   ?= test
SERVICE         ?= him-gateway
LDFLAGS         := -ldflags "-X main.Version=$(VERSION) -X main.Name=$(SERVICE)"
DOCKERFILE      ?= ./Dockerfile

all: proto wire tidy build

build:
	@mkdir -p bin
	$(GO) build $(LDFLAGS) -o bin/server ./cmd/server/

build-linux:
	@mkdir -p bin
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(GO) build $(LDFLAGS) -o bin/server ./cmd/server/

test:
	$(GO) test -race -count=1 ./...

test-short:
	$(GO) test -short -count=1 ./...

vet:
	$(GO) vet ./...

lint:
	golangci-lint run ./...

tidy:
	$(GO) mod tidy

clean:
	@rm -rf bin/
	$(GO) clean

proto:
	cd ../.. && $(PROTOC) -I . --go_out=. --go_opt=paths=source_relative services/gateway/internal/conf/conf.proto

wire:
	cd cmd/server && $(WIRE)

docker-build: build-linux
	docker buildx build --platform linux/amd64,linux/arm64 -t $(ACR_REGISTRY)/$(ACR_NAMESPACE)/$(SERVICE):$(IMAGE_VERSION) -f $(DOCKERFILE) --push .

docker-push:
	docker push $(ACR_REGISTRY)/$(ACR_NAMESPACE)/$(SERVICE):$(IMAGE_VERSION)

help:
	@awk '/^[a-zA-Z\-\_0-9]+:/ { helpMessage = match(lastLine, /^# (.*)/); if (helpMessage) { helpCommand = substr($$1, 0, index($$1, ":")); helpMessage = substr(lastLine, RSTART + 2, RLENGTH); printf "\033[36m%-22s\033[0m %s\n", helpCommand,helpMessage; } } { lastLine = $$0 }' $(MAKEFILE_LIST)

.DEFAULT_GOAL := help
