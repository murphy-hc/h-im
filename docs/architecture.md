# Architecture

## Service Map

```
Client (WS) ←→ Gateway ──gRPC──→ Auth / User / Message / Contact / Group / Chatroom / Media / Push
                    │                 │
                    └─── gRPC ───────┘ (inter-service calls, port 9200)
```

- **gateway**: WebSocket gateway — single client entry point. Exposes WS on 8080 for clients, gRPC on 9200 for inter-service calls. WS transport via `*server.WSServer` (wraps `net/http.Server` as Kratos transport). Follows full DDD layout including `data/` layer.
- **sequence**: Segment-based ID allocator using MySQL. Callers request a range `[start, end]` and hand out IDs locally. Atomic allocation via `SELECT FOR UPDATE`.
- **message**: Core message pipeline — persist, publish to Kafka, read receipts, replies.
- **user / auth / contact / group / chatroom / push / media**: gRPC services, each with Kratos observability stack.

## Message Flow

```
Sender → Gateway → Message Service → Sequence (get ID segment)
                                   → MySQL (persist)
                                   → Kafka (publish)
                                           ↓
Receiver ← Gateway ← Kafka consumer ←───┘
```

## Observability Stack (all services)

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

## DDD Layout

Every service follows this exact structure:

```
services/{name}/
├── Makefile
├── Dockerfile
├── go.mod
├── cmd/server/
│   ├── main.go                # Kratos app: config + tracing + metrics + wireApp
│   ├── wire.go                # wireinject: wireApp(bc, meter) → *kratos.App
│   └── wire_gen.go            # Wire-generated (do not edit)
├── configs/
│   └── config.yaml
└── internal/
    ├── conf/
    │   ├── conf.proto          # Bootstrap, Server, Data, Log, Otel
    │   └── conf.pb.go          # protoc-generated
    ├── server/
    │   ├── server.go           # GRPCProviderSet + HTTPProviderSet
    │   ├── grpc.go             # NewGRPCServer with middleware chain
    │   └── http.go             # NewHTTPServer → /metrics + /ping
    ├── service/
    │   ├── service.go          # ProviderSet: New{Name}Service
    │   └── {name}.go           # gRPC handler → biz.UseCase
    ├── biz/
    │   ├── biz.go              # ProviderSet: New{Name}UseCase
    │   ├── repo.go             # Repository interface
    │   └── {name}.go           # UseCase (domain logic)
    └── data/
        ├── data.go             # ProviderSet: NewData + New{Name}Repo
        └── {name}.go           # Repository (GORM)

Dependency direction: server → service → biz ← data
```

## Adding a gRPC Method

1. Define in `proto/him/{service}/v1/` and run `buf generate`
2. Add method to `internal/service/{name}.go` (delegates to `biz.UseCase`)
3. Add UseCase method in `internal/biz/{name}.go`
4. If data access needed: add interface method in `internal/biz/repo.go`, implement in `internal/data/{name}.go`
5. Run `wire ./cmd/server/` (or `make wire`) to regenerate

## Service Ports

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

## K8s Deployment

All services deployed on Alibaba Cloud ACK, namespace `default`. Service discovery via K8s CoreDNS.

### gRPC Client Dial

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
