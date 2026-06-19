.PHONY: all build build-linux test test-short vet lint tidy clean proto wire \
        docker-build docker-push help

# ── Variables ──────────────────────────────────────────────────────────────────
GO              ?= go
PROTOC          ?= protoc
WIRE            ?= wire
VERSION         ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
IMAGE_VERSION   ?= $(VERSION)
ACR_REGISTRY    ?= halei-acr-new-registry.eu-central-1.cr.aliyuncs.com
ACR_NAMESPACE   ?= test
SERVICE         ?= him-sequence

# go build -ldflags "-X main.Version=x.y.z"
LDFLAGS  := -ldflags "-X main.Version=$(VERSION) -X main.Name=$(SERVICE)"

DOCKERFILE ?= ./Dockerfile

# ── All ───────────────────────────────────────────────────────────────────────
# generate all
all: proto wire tidy build

# ── Build ─────────────────────────────────────────────────────────────────────
# build
build:
	@echo "Building $(SERVICE)..."
	@mkdir -p bin
	$(GO) build $(LDFLAGS) -o bin/server ./cmd/server/
	@echo "  ► bin/server"

# build for linux
build-linux:
	@echo "Building $(SERVICE) for linux amd64..."
	@mkdir -p bin
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(GO) build $(LDFLAGS) -o bin/server ./cmd/server/
	@echo "  ► bin/server"

# ── Test ───────────────────────────────────────────────────────────────────────
test:
	@echo "Running tests..."
	$(GO) test -race -count=1 ./...

test-short:
	@echo "Running tests (short)..."
	$(GO) test -short -count=1 ./...

# ── Vet / Lint ────────────────────────────────────────────────────────────────
vet:
	@echo "Running go vet..."
	$(GO) vet ./...

lint:
	@echo "Running golangci-lint..."
	golangci-lint run ./...

# ── Tidy ──────────────────────────────────────────────────────────────────────
tidy:
	@echo "Tidying..."
	$(GO) mod tidy

# ── Clean ─────────────────────────────────────────────────────────────────────
clean:
	@echo "Cleaning..."
	@rm -rf bin/
	$(GO) clean

# ── Proto ─────────────────────────────────────────────────────────────────────
proto:
	@echo "Generating conf.pb.go..."
	cd ../.. && $(PROTOC) -I . \
		--go_out=. --go_opt=paths=source_relative \
		services/sequence/internal/conf/conf.proto
	@echo "  ► internal/conf/conf.pb.go"

# ── Wire ──────────────────────────────────────────────────────────────────────
wire:
	@echo "Generating wire_gen.go..."
	cd cmd/server && $(WIRE)
	@echo "  ► cmd/server/wire_gen.go"

# ── Docker ────────────────────────────────────────────────────────────────────
# build docker image
docker-build: build-linux
	docker buildx build --platform linux/amd64,linux/arm64 \
		-t $(ACR_REGISTRY)/$(ACR_NAMESPACE)/$(SERVICE):$(IMAGE_VERSION) \
		-f $(DOCKERFILE) --push .

# push docker image
docker-push:
	docker push $(ACR_REGISTRY)/$(ACR_NAMESPACE)/$(SERVICE):$(IMAGE_VERSION)

# ── Help ──────────────────────────────────────────────────────────────────────
# show help
help:
	@echo ''
	@echo 'Usage:'
	@echo ' make [target]'
	@echo ''
	@echo 'Targets:'
	@awk '/^[a-zA-Z\-\_0-9]+:/ { \
		helpMessage = match(lastLine, /^# (.*)/); \
			if (helpMessage) { \
				helpCommand = substr($$1, 0, index($$1, ":")); \
				helpMessage = substr(lastLine, RSTART + 2, RLENGTH); \
				printf "\033[36m%-22s\033[0m %s\n", helpCommand,helpMessage; \
			} \
		} \
		{ lastLine = $$0 }' $(MAKEFILE_LIST)

.DEFAULT_GOAL := help
