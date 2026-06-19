# Sequence 服务号段模式设计

## Context

当前 sequence 服务使用雪花算法生成非连续 ID，需改造为号段模式，支持连续自增 ID 和批量分配。

## Requirements

1. **连续自增 ID**：每条序列独立自增，ID 严格连续
2. **号段批量分配**：调用方获取号段区间 `[start, end]`，本地缓存后按步长依次使用
3. **命名序列**：不同业务类型独立序列，通过 `key` 标识（如 `message_id`、`user_id`）
4. **按需创建**：首次使用时自动创建序列，默认 `start=1, step=1, segment_size=100`
5. **高 QPS 支撑**：PG 直接号段分配，号段模式将调用方请求 QPS 降低 100x+

## Design

### Architecture

```
caller ──gRPC──→ Sequence Service ──SQL──→ PostgreSQL (sequences table)
                      │
                      └── NextBatchID(key, size?)
                              │
                              ▼
                     UPDATE sequences
                     SET next_val = next_val + step * size
                     WHERE key = $1
                     RETURNING (next_val - step*size) AS start,
                               (next_val - step) AS end,
                               step
```

Key: 单行 UPDATE 原子操作，PG 行锁保证不重不丢。

### Proto

```protobuf
syntax = "proto3";
package him.sequence.v1;

service SequenceService {
  rpc NextBatchID(NextBatchIDRequest) returns (NextBatchIDResponse);
}

message NextBatchIDRequest {
  string key  = 1;  // 序列名称
  int32  size = 2;  // 号段大小，<=0 则用默认值
}

message NextBatchIDResponse {
  int64 start = 1;  // 号段起始（含）
  int64 end   = 2;  // 号段结束（含）
  int32 step  = 3;  // 步长
}
```

### Database Schema

```sql
CREATE TABLE IF NOT EXISTS sequences (
    key          VARCHAR(128) PRIMARY KEY,
    next_val     BIGINT NOT NULL DEFAULT 1,
    step         INT    NOT NULL DEFAULT 1,
    segment_size INT    NOT NULL DEFAULT 100
);
```

### Core Logic

```
AllocateSegment(ctx, key, size):
  1. size <= 0 → size = 默认 segment_size
  2. INSERT INTO sequences (key) VALUES ($1) ON CONFLICT DO NOTHING  -- 自动初始化
  3. UPDATE sequences
     SET next_val = next_val + step * size
     WHERE key = $1
     RETURNING (next_val - step*size) AS start, (next_val - step) AS end, step
  4. 返回 {start, end, step}
```

### Caller Usage Pattern

调用方本地实现 `IDAllocator`：

```
type IDAllocator { seqClient, cache{start,end,current,step}, mutex }

NextID():
  lock
  if current > end:
    resp = seqClient.NextBatchID(key, defaultSize)
    cache = {start: resp.start, end: resp.end, step: resp.step, current: resp.start}
  id = current
  current += step
  unlock
  return id
```

### Layer Changes

| Layer | Action | Files |
|-------|--------|-------|
| `proto/` | 简化为只保留 NextBatchID RPC | `sequence.proto` |
| `internal/conf/` | 不变 | `conf.go` |
| `internal/data/` | **新增**：PG 连接 + SequenceRepo 实现 | `data.go`, `sequence.go` |
| `internal/biz/` | 替换 snowflake 为号段分配 + 定义 Repo 接口 | `biz.go`, `repo.go`, `sequence.go` |
| `internal/service/` | 适配新 proto 和 biz | `service.go`, `sequence.go` |
| `internal/server/` | 适配新 service 签名 | `grpc.go`, `server.go` |
| `cmd/server/` | 加入 data.ProviderSet，重写 main.go | `wire.go`, `main.go` |
| `pkg/snowflake/` | 不再被 sequence 使用（保留） | — |

### Error Handling

- 首次调用 key 不存在 → 自动 INSERT ON CONFLICT DO NOTHING（幂等）
- PG 连接失败 → 返回 gRPC Internal error
- 同一 key 并发请求 → PG 行锁串行化，天然安全

### Testing Strategy

- **Unit**: biz 层用 mock repo 测试号段分配逻辑
- **Unit**: data 层用 test PG 实例测试 SQL
- **Integration**: gRPC 端到端测试

### Acceptance Criteria

1. NextBatchID 返回正确号段区间，多次调用严格递增无重叠
2. 不同 key 的序列完全独立
3. size ≤ 0 时使用默认 segment_size
4. 并发请求同一 key 无竞态，不重不漏
