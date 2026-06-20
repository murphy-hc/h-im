# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

> **⚠️ 约束：禁止自动执行 `git add` 和 `git commit`。** 所有 git 操作需用户明确指示。
> 
> **⚠️ 约束：services/ 下所有微服务必须遵循 Kratos DDD 分层**，目录结构统一为 `cmd/` `internal/server/` `internal/service/` `internal/biz/` `internal/data/` `internal/conf/`，依赖方向 `server → service → biz ← data`。

## Project

Go monorepo for an IM (Instant Messaging) system. Module root: `github.com/murphy-hc/h-im`. All services use the Kratos v2 framework with Wire DI, OpenTelemetry tracing, Prometheus metrics, and GORM for database access.

## Tech Stack

| Component | Technology |
|-----------|-----------|
| Language | Go 1.24 |
| Framework | Kratos v2 (`github.com/go-kratos/kratos/v2`) |
| Inter-service | gRPC + Protobuf (Buf for shared protos, protoc for service conf) |
| DI | Google Wire (compile-time) |
| Database | MySQL via GORM (`gorm.io/gorm` + `gorm.io/driver/mysql`) |
| Cache | Redis |
| Message Queue | Kafka |
| Observability | OpenTelemetry (tracing) + Prometheus (metrics) + structured logging |
| Container | Docker (debian:bookworm-slim base image) |
| Registry | Alibaba Cloud ACR |

## Common Commands

### Root level

```bash
make build               # Build all services (go work sync → per-service make)
make test                # Run all tests
make tidy                # go mod tidy in all modules
make proto               # buf lint + buf generate (api protos)
make proto-conf          # protoc generate all services' conf.pb.go
make build-{name}        # Build a single service (e.g. make build-sequence)
make docker-up           # Start MySQL + Redis + Kafka locally
```

### Per-service (any service under services/)

```bash
cd services/{name}
make all                 # proto → wire → tidy → build
make build               # Compile to bin/server
make build-linux         # Cross-compile for Linux amd64
make test                # Run tests with race detector
make vet                 # go vet
make proto               # protoc generate conf.pb.go
make wire                # Wire generate wire_gen.go
make docker-build        # Build Linux binary + multi-arch Docker image + push
make help                # Show all targets
```

### Development workflow

```bash
# After changing conf.proto:
cd services/{name} && make proto && make wire && make build

# After changing wire.go or any constructor:
cd services/{name} && make wire && make build

# Build single service with version info:
cd services/{name} && go build -ldflags "-X main.Version=1.0.0" -o bin/server ./cmd/server/
```

## Architecture

### Service Map

```
Client (WS) ←→ Gateway ──gRPC──→ Auth / User / Message / Contact / Group / Chatroom / Media / Push
                                      │
                                  Sequence (segment-based ID allocation)
```

- **gateway**: WebSocket gateway — the single client entry point. Handles connections, auth, message routing. Returns `*server.WSServer` (wraps `net/http.Server` as Kratos transport).
- **sequence**: Segment-based ID allocator using MySQL. Callers request a range `[start, end]` and hand out IDs locally. Atomic allocation via `SELECT FOR UPDATE`.
- **message**: Core message pipeline — persist, publish to Kafka, read receipts, replies.
- **user / auth / contact / group / chatroom / push / media**: gRPC services, each with Kratos observability stack.

### Message Flow

```
Sender → Gateway → Message Service → Sequence (get ID segment)
                                   → MySQL (persist)
                                   → Kafka (publish)
                                           ↓
Receiver ← Gateway ← Kafka consumer ←───┘
```

### Observability Stack (all services)

```
main.go:
  1. flag.Parse() → config.Load() → conf.Bootstrap.Scan()
  2. tracing.InitTracer(bc.Otel) → defer shutdown   (OTLP HTTP/gRPC exporter)
  3. metrics.NewPrometheusMeter(name, env)            (Prometheus exporter)
  4. kratos structured logger with trace.id + span.id (filtered by log.level)
  5. wireApp(&bc, meter) → app.Run()
```

gRPC middleware chain: `recovery → tracing → metadata → metrics`

HTTP endpoints: `/metrics` (Prometheus), `/ping` (health)

## Kratos DDD Layout

Every service (except gateway) follows this exact structure:

```
services/{name}/
├── Makefile                   # build, proto, wire, docker, test targets
├── Dockerfile                 # debian:bookworm-slim, COPY bin/ + configs/
├── go.mod
├── cmd/server/
│   ├── main.go                # Kratos app: config + tracing + metrics + wireApp
│   ├── wire.go                # wireinject: wireApp(bc, meter) → *kratos.App
│   └── wire_gen.go            # Wire-generated (do not edit)
├── configs/
│   └── config.yaml            # Structured: server.{grpc,http}, otel, log, data
└── internal/
    ├── conf/
    │   ├── conf.proto          # Bootstrap, Server, Data, Log, Otel definition
    │   └── conf.pb.go          # protoc-generated
    ├── server/
    │   ├── server.go           # GRPCProviderSet + HTTPProviderSet (with Wire FieldsOf)
    │   ├── grpc.go             # NewGRPCServer(bc, meter, svc) with middleware chain
    │   └── http.go             # NewHTTPServer(bc, meter) → /metrics + /ping
    ├── service/
    │   ├── service.go          # ProviderSet: New{Name}Service
    │   └── {name}.go           # gRPC handler — calls biz.UseCase
    ├── biz/
    │   ├── biz.go              # ProviderSet: New{Name}UseCase
    │   ├── repo.go             # Repository interface definition
    │   └── {name}.go           # UseCase implementation (domain logic)
    └── data/
        ├── data.go             # ProviderSet: NewData(bc) + New{Name}Repo
        └── {name}.go           # Repository implementation (GORM)
```

**Gateway exception**: uses `server.WSServer` (wraps `net/http.Server` as Kratos transport) instead of gRPC. No `data/` layer.

### Dependency Direction

```
server → service → biz ← data
```

### Adding a New gRPC Method

1. Define in `proto/him/{service}/v1/` and run `buf generate`
2. Add method to `internal/service/{name}.go` (delegates to `biz.UseCase`)
3. Add UseCase method in `internal/biz/{name}.go`
4. If data access needed: add interface method in `internal/biz/repo.go`, implement in `internal/data/{name}.go`
5. Run `wire ./cmd/server/` (or `make wire`) to regenerate

## Shared Packages (`pkg/`)

| Package | Purpose |
|---------|---------|
| `pkg/database` | GORM connection factory (MySQL driver) |
| `pkg/tracing` | OTLP trace exporter (HTTP/gRPC, configurable timeout/transport) |
| `pkg/metrics` | Prometheus meter (`NewMeterProvider` accepts any `sdkmetric.Reader`) |
| `pkg/errcode` | Domain error code constants |
| `pkg/jwt` | JWT issue/validate helpers |
| `pkg/pagination` | Pagination request/response utilities |
| `pkg/redis` | Redis client wrapper |
| `pkg/kafka` | Kafka producer/consumer stub |
| `pkg/logger` | slog-based logger factory |

## Module Conventions

- Module path: `github.com/murphy-hc/h-im/{submodule}`
- API protos: `proto/him/{service}/v1/` → `gen/go/him/{service}/v1/` (via Buf)
- Service config protos: `services/{name}/internal/conf/conf.proto` → `conf.pb.go` (via protoc)
- Shared packages: `pkg/` — imported by services, never the reverse
- Wire DI: all services use `wireApp(bc *conf.Bootstrap, meter metric.Meter) (*kratos.App, func(), error)`
- Config: every service loads from `configs/` directory via `-conf` flag

## K8s Deployment

All services deployed on Alibaba Cloud ACK, namespace `default`. Service discovery via K8s CoreDNS.

### Service Addresses (gRPC Client Dial)

All inter-service gRPC calls MUST use Kratos `transport/grpc` client with DNS resolver + round_robin:

```go
import "github.com/go-kratos/kratos/v2/transport/grpc"

conn, _ := grpc.DialInsecure(
    context.Background(),
    grpc.WithEndpoint("discovery:///sequence.default.svc.cluster.local:9108"),
)
client := pb.NewSequenceClient(conn)
```

- `discovery:///` prefix enables Kratos client-side load balancing via DNS resolver
- `grpc.DialInsecure` wraps Kratos middleware (tracing, recovery) automatically

### Deploy Configuration

Each service has `deploy/k8s/{service}/`:
- `service.yaml` — ClusterIP service (gRPC + HTTP ports)
- `deployment.yaml` — Deployment with liveness/readiness probes on `/ping`
- `ingress.yaml` — ALB Ingress (gateway only, exposes `/ws` and `/metrics`)

### Image Registry

```
halei-acr-new-registry.eu-central-1.cr.aliyuncs.com/test/him-{service}:{version}
```

### Service Ports

| Service | gRPC | HTTP (metrics) |
|---------|:----:|:---:|
| auth | 9100 | 8100 |
| user | 9101 | 8101 |
| message | 9102 | 8102 |
| contact | 9103 | 8103 |
| group | 9104 | 8104 |
| chatroom | 9105 | 8105 |
| push | 9106 | 8106 |
| media | 9107 | 8107 |
| sequence | 9108 | 8108 |
| gateway | — (WS 8080, gRPC 9200) | 8180 |
