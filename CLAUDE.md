# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project

Go monorepo for an IM (Instant Messaging) system. Module root: `github.com/murphy-hc/h-im`. Each service under `services/` is an independent Go module wired together via `go.work`.

## Tech Stack

| Component | Technology |
|-----------|-----------|
| Language | Go 1.24 |
| Inter-service | gRPC + Protobuf |
| Proto management | Buf |
| Primary DB | PostgreSQL |
| Cache / State | Redis |
| Message Queue | RocketMQ |
| DI | Google Wire (compile-time) |

## Common Commands

```bash
make build          # Build all services
make lint           # go vet + golangci-lint
make test           # Run all tests with race detector
make tidy           # go mod tidy in all modules
make work-sync      # go work sync
make proto          # buf lint + buf generate
make docker-up      # Start PG + Redis + RocketMQ locally
make docker-down    # Stop local infra
```

## Architecture

### Service Map

```
Client (WS) ←→ Gateway ──gRPC──→ Auth / User / Message / Contact / Group / Chatroom / Media / Push
                                      │
                                  Sequence (ID generation)
```

- **gateway**: WebSocket gateway — single entry point for clients. Handles connection lifecycle, auth, and message routing.
- **sequence**: Distributed ID generator (snowflake). Called by other services for unique message/sequence IDs.
- **message**: Core message pipeline — persist to PostgreSQL, publish to RocketMQ for delivery, handle read receipts and replies.
- **user**: Registration, login, profile.
- **auth**: Token issuance and validation, permissions.
- **contact**: Friend relationships, blacklists.
- **group**: Group creation, membership, settings.
- **chatroom**: Priority-based chat room messaging (high/medium/low priority queues).
- **push**: Offline push notifications (APNs, FCM, vendor channels).
- **media**: File/image/audio/video upload and download.

### Message Flow

```
Sender → Gateway → Message Service → Sequence (get ID)
                                   → PostgreSQL (persist)
                                   → RocketMQ (publish)
                                           ↓
Receiver ← Gateway ← RocketMQ consumer ←───┘
```

### Priority Messages (Chatroom)

Each priority level gets its own RocketMQ queue. Consumer polls high-priority first. Message ordering is guaranteed within the same priority level via RocketMQ MessageGroup.

### Deduplication

Every message carries a snowflake `message_id`. Consumer uses Redis `SETNX key:msg:dedup:{message_id}` with 7-day TTL.

## Kratos DDD Layout

Every service follows the Kratos-inspired DDD layered architecture:

```
services/{name}/
├── go.mod
├── cmd/server/
│   ├── main.go           # Load config → wireApp() → Serve
│   ├── wire.go           # Wire injector declaration (//go:build wireinject)
│   └── wire_gen.go       # Wire-generated (do not edit)
├── configs/
│   └── config.yaml       # Local development config
└── internal/
    ├── server/            # Transport layer — gRPC/HTTP server creation
    │   ├── server.go      #   ProviderSet
    │   └── grpc.go        #   NewGRPCServer (or ws.go for gateway)
    ├── service/           # Application layer — gRPC handler implementations
    │   ├── service.go     #   ProviderSet
    │   └── {name}.go      #   {Name}Service struct & gRPC methods
    ├── biz/               # Domain layer — business logic, repository interfaces
    │   ├── biz.go         #   ProviderSet
    │   ├── repo.go        #   Repository interface definition
    │   └── {name}.go      #   UseCase implementation
    ├── data/              # Infrastructure layer — repository implementations
    │   ├── data.go        #   ProviderSet + Data aggregate (DB/Redis/MQ clients)
    │   └── {name}.go      #   Repository implementation
    └── conf/              # Configuration structs
        └── conf.go
```

### Dependency Direction

```
server → service → biz ← data
```

- **biz/** defines interfaces only — no imports of ORM, Redis driver, or other infra.
- **data/** implements `biz.Repo` interfaces, handles PO ↔ Domain mapping.
- **service/** converts Proto DTOs ↔ biz models, calls UseCases. No business logic.
- **server/** wires gRPC server with services. Separated from main.go for testability.

### Adding a New gRPC Method

1. Define in `proto/him/{service}/v1/` and run `buf generate`
2. Add method to `internal/service/{name}.go` (delegates to `biz.UseCase`)
3. Add UseCase method in `internal/biz/{name}.go`
4. If data access needed: add interface method in `internal/biz/repo.go`, implement in `internal/data/{name}.go`
5. Run `wire ./cmd/server/` to regenerate `wire_gen.go`

## Module Conventions

- Module path: `github.com/murphy-hc/h-im/{submodule}`
- Proto: `proto/him/{service}/v1/` → generated to `gen/go/him/{service}/v1/`
- Shared packages: `pkg/` — imported by services, never the reverse
- gRPC ports: 9100 range (see `.env.example`)
- Gateway is the only non-gRPC service (WebSocket on :8080)

## Build & DI

- `cd services/{name} && go build ./cmd/server/` — build one service
- `wire ./cmd/server/` — regenerate DI code after changing constructors
- After adding/removing imports, run `go mod tidy` in the affected service
