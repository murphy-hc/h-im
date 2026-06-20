# 多应用隔离与鉴权设计

## Context

IM 系统需要支持多个业务应用接入，不同应用之间数据隔离。每个应用分配 `app_id` + `app_secret`，客户端在 WebSocket 连接时携带鉴权参数，Gateway 在握手阶段完成鉴权。鉴权通过后直接进入消息通信，无需二次登录。

## Requirements

1. **App 管理**：每个业务应用有独立 `app_id` + `app_secret`，存储在 PostgreSQL `apps` 表中，支持动态增删
2. **连接级鉴权**：客户端通过 WebSocket URL 参数传递 `app_id` + `user_id` + `token`，Gateway 在 upgrade 前校验
3. **Token 签发**：由业务应用后端负责，IM 系统只校验。算法：`Base64(JSON{signature: SHA1(appSecret+userID+curTime+ttl), curTime, ttl})`
4. **Proto 清理**：移除 `AuthRequest`/`AuthResponse` 消息和 `FRAME_TYPE_AUTH_REQ`/`FRAME_TYPE_AUTH_RESP` 枚举值
5. **无二次登录**：鉴权成功后连接立即就绪，直接进入消息收发

## Design

### 连接流程

```
Client ── GET /ws?app_id=xxx&user_id=xxx&token=xxx&device_id=xxx ──→ Gateway
                                                                       │
                                                            ① 查 apps 表: SELECT * WHERE app_id=?
                                                            ② Base64 解码 token → {signature, curTime, ttl}
                                                            ③ 验签: SHA1(appSecret+userID+curTimeStr+ttlStr) == signature
                                                            ④ 检查过期: curTime+ttl > now_ms
                                                            ⑤ 鉴权失败 → 401 + 错误原因
                                                            ⑥ 鉴权通过 → WS Upgrade
                                                                       │
Client ←──────────────────────── WS 连接就绪 ─────────────────────────┘
  │                                                                    
  │  直接发送业务帧 (PRIVATE_CHAT / GROUP_CHAT / HEARTBEAT ...)
  │──────────────────────────────────────────────────────────────→
```

### 数据库

```sql
CREATE TABLE apps (
    id         BIGSERIAL PRIMARY KEY,
    app_id     VARCHAR(64) UNIQUE NOT NULL,
    app_secret VARCHAR(256) NOT NULL,
    app_name   VARCHAR(128),
    enabled    BOOLEAN DEFAULT true,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);
```

### Token 格式

业务应用后端签发：

```go
str := fmt.Sprintf("%s%s%d%d", appSecret, accountID, curTime, ttl)
signature := sha1(str)
token := base64(json.Marshal(map[string]interface{}{
    "signature": signature,
    "curTime":   curTime,
    "ttl":       ttl,
}))
```

### FrameType 变更

移除两个枚举值：

```diff
- FRAME_TYPE_AUTH_REQ     = 1;
- FRAME_TYPE_AUTH_RESP    = 2;
  FRAME_TYPE_PRIVATE_CHAT = 3;
  FRAME_TYPE_PRIVATE_ACK  = 4;
  ...
```

`ws.proto` 删除 `AuthRequest` / `AuthResponse` 消息。

### 文件变更

| 操作 | 文件 | 职责 |
|------|------|------|
| 修改 | `proto/him/gateway/v1/ws.proto` | 移除 AuthRequest/Response，FrameType 编号重排 |
| 新增 | `services/gateway/internal/data/app_repo.go` | GORM model + 查询 |
| 新增 | `services/gateway/internal/biz/auth.go` | `VerifyAppToken(appID, userID, token) error` |
| 修改 | `services/gateway/internal/service/gateway.go` | 读取 query params + 调鉴权 |
| 修改 | `services/gateway/internal/biz/gateway.go` | 移除 AUTH_REQ case |
| 修改 | `proto/him/gateway/v1/gateway.proto` | 不改（gRPC push 不变） |

### Testing Strategy

- **Unit**: `VerifyAppToken` 正确的 token 通过、错误的 secret 失败、过期失败
- **Integration**: WS 连接携带有效/无效参数的行为

### Acceptance Criteria

1. 有效 token 连接成功 → WS upgrade + 直接收发消息
2. 无效 app_id → 401 拒绝
3. 错误 token → 401 拒绝
4. 过期 token → 401 拒绝
5. Proto FrameType 不含 AUTH_REQ/AUTH_RESP
