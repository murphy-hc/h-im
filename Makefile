.PHONY: all build lint test clean proto proto-lint proto-gen tidy work-sync

# ── Go ────────────────────────────────────────────────────────────────────────
all: proto-gen tidy work-sync build

build:
	@echo "Building all services..."
	@go work sync
	@for svc in $$(ls -d services/*/); do \
		name=$$(basename $$svc); \
		echo "  ► $$name"; \
		(cd $$svc && go build ./cmd/server/) || exit 1; \
	done
	@echo "All services built."

lint:
	@echo "Running go vet..."
	@go vet ./...
	@echo "Running golangci-lint..."
	@golangci-lint run ./...

test:
	@echo "Running all tests..."
	@go test -race -short ./...

clean:
	@echo "Cleaning..."
	@for svc in $$(ls -d services/*/); do \
		(cd $$svc && go clean); \
	done

tidy:
	@echo "Tidying all modules..."
	@for svc in $$(ls -d services/*/); do \
		(cd $$svc && go mod tidy); \
	done

work-sync:
	@go work sync

# ── Proto ──────────────────────────────────────────────────────────────────────
proto-lint:
	buf lint

proto-gen:
	buf generate

proto: proto-lint proto-gen

# ── Docker ────────────────────────────────────────────────────────────────────
docker-up:
	docker compose -f docker/docker-compose.yml up -d

docker-down:
	docker compose -f docker/docker-compose.yml down

docker-reset:
	docker compose -f docker/docker-compose.yml down -v
	docker compose -f docker/docker-compose.yml up -d
