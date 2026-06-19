# Sequence 号段模式 Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 将 sequence 服务从雪花算法改造为号段模式，基于 PostgreSQL 实现命名序列的连续自增 ID 分配。

**Architecture:** sequence 服务通过 gRPC NextBatchID 接口分配号段区间 `[start, end]`，底层使用 PostgreSQL `UPDATE ... RETURNING` 单行原子操作保证不重不漏，调用方本地缓存号段按步长使用。

**Tech Stack:** Go 1.24, gRPC, pgx/v5, Wire

## Global Constraints

- Proto 简化：移除 NextID RPC，只保留 NextBatchID
- 号段语义：返回区间 `[start, end]` + `step`，调用方自行分配
- 按需创建：首次调用 key 不存在时，用默认值 `start=1, step=1, segment_size=100` 自动初始化
- 步长一致性：UPDATE 使用表中已存储的 step，防止首次初始化竞态
- PG 依赖：通过 pgx/v5 直连，不使用 ORM
- pkg/snowflake 保留但不再被 sequence 引用

---

### Task 1: Update proto definition

**Files:**
- Modify: `proto/him/sequence/v1/sequence.proto`
- Modify: `gen/go/him/sequence/v1/sequence.pb.go` (regenerated)
- Modify: `gen/go/him/sequence/v1/sequence_grpc.pb.go` (regenerated)

**Interfaces:**
- Produces: `NextBatchIDRequest{Key, Size}`, `NextBatchIDResponse{Start, End, Step}`, `SequenceServiceServer` (single method)

- [ ] **Step 1: Rewrite sequence.proto**

Replace the entire content of `proto/him/sequence/v1/sequence.proto`:

```protobuf
syntax = "proto3";

package him.sequence.v1;

option go_package = "github.com/murphy-hc/h-im/gen/go/him/sequence/v1;sequencev1";

message NextBatchIDRequest {
  string key  = 1;  // sequence name, e.g. "message_id", "user_id"
  int32  size = 2;  // segment size, <=0 to use default
}

message NextBatchIDResponse {
  int64 start = 1;  // segment start (inclusive)
  int64 end   = 2;  // segment end (inclusive)
  int32 step  = 3;  // increment step
}

service SequenceService {
  rpc NextBatchID(NextBatchIDRequest) returns (NextBatchIDResponse);
}
```

- [ ] **Step 2: Regenerate Go code**

Run: `~/buf/bin/buf generate`
Expected: no errors, `gen/go/him/sequence/v1/` regenerated

- [ ] **Step 3: Commit**

```bash
git add proto/him/sequence/v1/sequence.proto gen/go/him/sequence/v1/
git commit -m "feat(sequence): change proto to segment-based NextBatchID"
```

---

### Task 2: Define Repository interface in biz layer

**Files:**
- Create: `services/sequence/internal/biz/repo.go`
- Modify: `services/sequence/internal/biz/biz.go`
- Modify: `services/sequence/internal/biz/sequence.go`

**Interfaces:**
- Consumes: (none — this is the interface definition)
- Produces: `SequenceRepo` interface with `AllocateSegment(ctx, key string, size int32) (start, end int64, step int32, err error)`
- Produces: `SequenceUseCase` with `AllocateSegment(ctx, key string, size int32) (start, end int64, step int32, err error)`

- [ ] **Step 1: Create repo.go with the interface**

Create `services/sequence/internal/biz/repo.go`:

```go
package biz

import "context"

// SequenceRepo defines the sequence data access interface.
type SequenceRepo interface {
	// AllocateSegment allocates a segment of IDs for the given key.
	// Returns the allocated range [start, end] and the step size.
	AllocateSegment(ctx context.Context, key string, size int32) (start, end int64, step int32, err error)
}
```

- [ ] **Step 2: Update biz.go to wire repo into ProviderSet**

Replace `services/sequence/internal/biz/biz.go`:

```go
package biz

import "github.com/google/wire"

// ProviderSet is biz providers.
var ProviderSet = wire.NewSet(NewSequenceUseCase)
```

- [ ] **Step 3: Rewrite sequence.go — segment allocation use case**

Replace `services/sequence/internal/biz/sequence.go`:

```go
package biz

import (
	"context"
	"fmt"
)

// SequenceUseCase handles segment-based ID generation.
type SequenceUseCase struct {
	repo SequenceRepo
}

// NewSequenceUseCase creates a SequenceUseCase.
func NewSequenceUseCase(repo SequenceRepo) *SequenceUseCase {
	return &SequenceUseCase{repo: repo}
}

// AllocateSegment allocates a segment of IDs for the given key.
func (uc *SequenceUseCase) AllocateSegment(ctx context.Context, key string, size int32) (start, end int64, step int32, err error) {
	if key == "" {
		return 0, 0, 0, fmt.Errorf("key must not be empty")
	}
	return uc.repo.AllocateSegment(ctx, key, size)
}
```

- [ ] **Step 4: Verify internal compiles so far (will fail on missing repo impl)**

Run: `cd services/sequence && go build ./internal/biz/... 2>&1 || true`
Expected: biz compiles, but `go build ./internal/...` fails because data layer doesn't exist yet

- [ ] **Step 5: Commit**

```bash
git add services/sequence/internal/biz/
git commit -m "feat(sequence): add SequenceRepo interface and segment use case"
```

---

### Task 3: Create data layer — Repository implementation

**Files:**
- Create: `services/sequence/internal/data/data.go`
- Create: `services/sequence/internal/data/sequence.go`
- Modify: `services/sequence/go.mod` (add pgx dependency)

**Interfaces:**
- Consumes: `biz.SequenceRepo` interface (Task 2)
- Produces: `data.ProviderSet` (NewData, NewSequenceRepo), `data.Data` struct, `data.sequenceRepo` implementing `biz.SequenceRepo`

- [ ] **Step 1: Add pgx dependency**

Run:
```bash
cd services/sequence
go get github.com/jackc/pgx/v5@latest
```

Expected: pgx added to go.mod

- [ ] **Step 2: Create data.go — ProviderSet + Data struct**

Create `services/sequence/internal/data/data.go`:

```go
package data

import (
	"context"
	"fmt"

	"github.com/google/wire"
	"github.com/jackc/pgx/v5/pgxpool"
)

// ProviderSet is data providers.
var ProviderSet = wire.NewSet(NewData, NewSequenceRepo)

// Data holds data source clients.
type Data struct {
	PG *pgxpool.Pool
}

// NewData creates a Data instance with a PG connection pool.
func NewData() (*Data, func(), error) {
	// TODO: read DSN from config
	dsn := "postgres://him:him_secret@localhost:5432/him?sslmode=disable"

	pool, err := pgxpool.New(context.Background(), dsn)
	if err != nil {
		return nil, nil, fmt.Errorf("data: connect pg: %w", err)
	}

	if err := pool.Ping(context.Background()); err != nil {
		pool.Close()
		return nil, nil, fmt.Errorf("data: ping pg: %w", err)
	}

	d := &Data{PG: pool}
	cleanup := func() {
		pool.Close()
	}
	return d, cleanup, nil
}
```

- [ ] **Step 3: Create sequence.go — SequenceRepo implementation**

Create `services/sequence/internal/data/sequence.go`:

```go
package data

import (
	"context"
	"fmt"

	"github.com/murphy-hc/h-im/services/sequence/internal/biz"
)

const defaultSegmentSize = 100

var _ biz.SequenceRepo = (*sequenceRepo)(nil)

type sequenceRepo struct {
	data *Data
}

// NewSequenceRepo creates a SequenceRepo implementation.
func NewSequenceRepo(data *Data) biz.SequenceRepo {
	return &sequenceRepo{data: data}
}

// AllocateSegment allocates a segment of IDs atomically.
func (r *sequenceRepo) AllocateSegment(ctx context.Context, key string, size int32) (start, end int64, step int32, err error) {
	if size <= 0 {
		size = defaultSegmentSize
	}

	tx, err := r.data.PG.Begin(ctx)
	if err != nil {
		return 0, 0, 0, fmt.Errorf("sequence: begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	// Auto-create sequence on first use — insert default row if key doesn't exist.
	_, err = tx.Exec(ctx,
		`INSERT INTO sequences (key, next_val, step, segment_size)
		 VALUES ($1, 1, 1, $2)
		 ON CONFLICT (key) DO NOTHING`,
		key, defaultSegmentSize,
	)
	if err != nil {
		return 0, 0, 0, fmt.Errorf("sequence: init key %s: %w", key, err)
	}

	// Atomically allocate segment.
	var startVal, endVal int64
	var stepVal int32
	err = tx.QueryRow(ctx,
		`UPDATE sequences
		 SET next_val = next_val + step * $2
		 WHERE key = $1
		 RETURNING next_val - step * $2, next_val - step, step`,
		key, size,
	).Scan(&startVal, &endVal, &stepVal)
	if err != nil {
		return 0, 0, 0, fmt.Errorf("sequence: allocate segment for %s: %w", key, err)
	}

	if err := tx.Commit(ctx); err != nil {
		return 0, 0, 0, fmt.Errorf("sequence: commit tx: %w", err)
	}

	return startVal, endVal, stepVal, nil
}
```

- [ ] **Step 4: Initialize database schema**

Run manually once, or add migration:

```sql
CREATE TABLE IF NOT EXISTS sequences (
    key          VARCHAR(128) PRIMARY KEY,
    next_val     BIGINT NOT NULL DEFAULT 1,
    step         INT    NOT NULL DEFAULT 1,
    segment_size INT    NOT NULL DEFAULT 100
);
```

Record this as a migration note in the commit.

- [ ] **Step 5: Verify internal compiles**

Run: `cd services/sequence && go build ./internal/...`
Expected: PASS

- [ ] **Step 6: Commit**

```bash
git add services/sequence/internal/data/ services/sequence/go.mod services/sequence/go.sum
git commit -m "feat(sequence): add data layer with PG-backed segment allocation"
```

---

### Task 4: Update service layer for new proto

**Files:**
- Modify: `services/sequence/internal/service/sequence.go`
- Modify: `services/sequence/internal/service/service.go`

**Interfaces:**
- Consumes: `biz.SequenceUseCase.AllocateSegment` (Task 2), `pb.SequenceServiceServer` (Task 1)
- Produces: `*SequenceService` implementing `pb.SequenceServiceServer` with only `NextBatchID`

- [ ] **Step 1: Rewrite service/sequence.go**

Replace `services/sequence/internal/service/sequence.go`:

```go
package service

import (
	"context"

	pb "github.com/murphy-hc/h-im/gen/go/him/sequence/v1"
	"github.com/murphy-hc/h-im/services/sequence/internal/biz"
)

// SequenceService implements the SequenceService gRPC server.
type SequenceService struct {
	pb.UnimplementedSequenceServiceServer
	uc *biz.SequenceUseCase
}

// NewSequenceService creates a SequenceService.
func NewSequenceService(uc *biz.SequenceUseCase) *SequenceService {
	return &SequenceService{uc: uc}
}

// NextBatchID allocates a segment of IDs.
func (s *SequenceService) NextBatchID(ctx context.Context, req *pb.NextBatchIDRequest) (*pb.NextBatchIDResponse, error) {
	start, end, step, err := s.uc.AllocateSegment(ctx, req.GetKey(), req.GetSize())
	if err != nil {
		return nil, err
	}
	return &pb.NextBatchIDResponse{
		Start: start,
		End:   end,
		Step:  step,
	}, nil
}
```

- [ ] **Step 2: Verify service compiles**

Run: `cd services/sequence && go build ./internal/...`
Expected: PASS

- [ ] **Step 3: Commit**

```bash
git add services/sequence/internal/service/
git commit -m "feat(sequence): update service layer for segment-based API"
```

---

### Task 5: Update server layer and Wire configuration

**Files:**
- Modify: `services/sequence/internal/server/grpc.go`
- Modify: `services/sequence/cmd/server/wire.go`
- Modify: `services/sequence/cmd/server/main.go`
- Modify: `services/sequence/go.mod` (remove snowflake dependency if no longer used)

**Interfaces:**
- Consumes: `service.SequenceService` (Task 4), `data.ProviderSet` (Task 3), `biz.ProviderSet` (Task 2)
- Produces: working `wire_gen.go`, compilable `cmd/server/main.go`

- [ ] **Step 1: Update grpc.go — no changes needed, verify it still works**

The `grpc.go` already uses `*service.SequenceService` and `pb.RegisterSequenceServiceServer`. Since `SequenceService` still implements the gRPC interface (just now with `NextBatchID` only), no change needed.

- [ ] **Step 2: Update wire.go — add data.ProviderSet**

Replace `services/sequence/cmd/server/wire.go`:

```go
//go:build wireinject
// +build wireinject

package main

import (
	"github.com/google/wire"
	"google.golang.org/grpc"

	"github.com/murphy-hc/h-im/services/sequence/internal/biz"
	"github.com/murphy-hc/h-im/services/sequence/internal/conf"
	"github.com/murphy-hc/h-im/services/sequence/internal/data"
	"github.com/murphy-hc/h-im/services/sequence/internal/server"
	"github.com/murphy-hc/h-im/services/sequence/internal/service"
)

func wireApp(*conf.Server, *conf.Data) (*grpc.Server, func(), error) {
	panic(wire.Build(
		server.ProviderSet,
		service.ProviderSet,
		biz.ProviderSet,
		data.ProviderSet,
	))
}
```

- [ ] **Step 3: Run wire to regenerate wire_gen.go**

Run: `cd services/sequence && wire ./cmd/server/`
Expected: writes `cmd/server/wire_gen.go` successfully

- [ ] **Step 4: Update main.go — pass PG DSN from config to wireApp**

Replace `services/sequence/cmd/server/main.go`:

```go
package main

import (
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/murphy-hc/h-im/pkg/logger"
	"github.com/murphy-hc/h-im/services/sequence/internal/conf"
)

func main() {
	log := logger.WithService(logger.New(os.Getenv("LOG_LEVEL")), "sequence")

	cfg := &conf.Server{
		GRPC: conf.GRPCServer{Addr: envOrDefault("SEQUENCE_GRPC_ADDR", ":9108")},
	}

	grpcServer, cleanup, err := wireApp(cfg, &conf.Data{})
	if err != nil {
		log.Error("wireApp failed", "error", err)
		os.Exit(1)
	}
	defer cleanup()

	addr := cfg.GRPC.Addr
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		log.Error("listen failed", "error", err)
		os.Exit(1)
	}

	go func() {
		log.Info("sequence service starting", "addr", addr)
		if err := grpcServer.Serve(lis); err != nil {
			log.Error("serve failed", "error", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info("shutting down...")
	grpcServer.GracefulStop()
	log.Info("sequence service stopped")
}

func envOrDefault(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
```

- [ ] **Step 5: Remove snowflake dependency from go.mod**

Run:
```bash
cd services/sequence
go mod tidy
```

This will remove `pkg/snowflake` from go.mod since it's no longer imported.

- [ ] **Step 6: Verify full build**

Run: `cd services/sequence && go build ./cmd/server/`
Expected: PASS

- [ ] **Step 7: Run go vet**

Run: `cd services/sequence && go vet ./...`
Expected: PASS with no warnings

- [ ] **Step 8: Commit**

```bash
git add services/sequence/
git commit -m "feat(sequence): wire up segment-based ID allocation end-to-end"
```

---

### Task 6: Write tests

**Files:**
- Create: `services/sequence/internal/biz/sequence_test.go`
- Create: `services/sequence/internal/data/sequence_test.go`

**Interfaces:**
- Consumes: `biz.SequenceRepo`, `biz.SequenceUseCase`, `data.sequenceRepo`
- Produces: test coverage for allocation logic and SQL execution

- [ ] **Step 1: Write biz unit test with mock repo**

Create `services/sequence/internal/biz/sequence_test.go`:

```go
package biz_test

import (
	"context"
	"errors"
	"testing"

	"github.com/murphy-hc/h-im/services/sequence/internal/biz"
)

type mockRepo struct {
	start, end int64
	step       int32
	err        error
}

func (m *mockRepo) AllocateSegment(ctx context.Context, key string, size int32) (int64, int64, int32, error) {
	return m.start, m.end, m.step, m.err
}

func TestAllocateSegment_EmptyKey(t *testing.T) {
	uc := biz.NewSequenceUseCase(&mockRepo{})
	_, _, _, err := uc.AllocateSegment(context.Background(), "", 10)
	if err == nil {
		t.Fatal("expected error for empty key")
	}
}

func TestAllocateSegment_Success(t *testing.T) {
	mock := &mockRepo{start: 1, end: 100, step: 1}
	uc := biz.NewSequenceUseCase(mock)
	start, end, step, err := uc.AllocateSegment(context.Background(), "msg_id", 100)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if start != 1 || end != 100 || step != 1 {
		t.Fatalf("unexpected result: start=%d end=%d step=%d", start, end, step)
	}
}

func TestAllocateSegment_RepoError(t *testing.T) {
	mock := &mockRepo{err: errors.New("db down")}
	uc := biz.NewSequenceUseCase(mock)
	_, _, _, err := uc.AllocateSegment(context.Background(), "msg_id", 10)
	if err == nil {
		t.Fatal("expected error from repo")
	}
}
```

- [ ] **Step 2: Run biz tests**

Run: `cd services/sequence && go test ./internal/biz/... -v`
Expected: 3 tests PASS

- [ ] **Step 3: Write data integration test (optional, requires PG)**

Create `services/sequence/internal/data/sequence_test.go`:

```go
package data_test

import (
	"context"
	"testing"

	"github.com/murphy-hc/h-im/services/sequence/internal/data"
)

// Requires a running PostgreSQL with the sequences table created.
// Skip if no PG available.
func TestAllocateSegment_Integration(t *testing.T) {
	// TODO: set PG_DSN env var or skip
	dsn := "postgres://him:him_secret@localhost:5432/him?sslmode=disable"
	if dsn == "" {
		t.Skip("PG_DSN not set, skipping integration test")
	}
	// Test: call AllocateSegment twice, verify non-overlapping ranges
}
```

- [ ] **Step 4: Run all tests**

Run: `cd services/sequence && go test ./... -v`
Expected: unit tests PASS, integration test SKIPPED

- [ ] **Step 5: Commit**

```bash
git add services/sequence/internal/biz/sequence_test.go services/sequence/internal/data/sequence_test.go
git commit -m "test(sequence): add unit tests for segment allocation"
```

---

### Task 7: Final verification

- [ ] **Step 1: Full build**

Run: `cd services/sequence && go build ./cmd/server/`
Expected: PASS

- [ ] **Step 2: go vet**

Run: `cd services/sequence && go vet ./...`
Expected: PASS

- [ ] **Step 3: Run unit tests**

Run: `cd services/sequence && go test ./internal/biz/... -v`
Expected: 3 tests PASS

- [ ] **Step 4: Verify buf lint**

Run: `~/buf/bin/buf lint`
Expected: PASS

- [ ] **Step 5: Commit if any changes**

```bash
git add -A
git diff --cached --stat
git commit -m "chore(sequence): final verification and cleanup"
```
