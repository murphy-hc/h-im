# Heartbeat & Online Status Design

## Overview

实现心跳和用户在线状态维护，支持多 gateway 实例间的消息路由。

## Requirements

1. 客户端每 10 秒发送一次心跳，连续 3 分钟没有收到过心跳 → 判定离线
2. 每次收到客户端心跳回复一条 echo，echo 发送失败记为一次心跳中断
3. 消息投递时判断用户在线状态：离线走 PullMessages 拉取；在线则直连对应 gateway 实例投递

## Data Model

### Memory (per Gateway Instance)

```go
type connState struct {
    Conn                    *websocket.Conn
    LastSuccessHeartbeat    time.Time
    ConsecutiveEchoFailures int
}
```

`localConns` 类型从 `map[userID]map[deviceID]*websocket.Conn` 改为 `map[userID]map[deviceID]*connState`。

### Redis Keys

| Key | Type | Value |
|-----|------|-------|
| `user:online:{userID}:{deviceID}` | Hash | `instance_id`, `gateway_addr`, `last_heartbeat_ts` |
| `user:devices:{userID}` | Set | 在线 deviceID 集合 |

### Config (gateway conf.proto)

```protobuf
message Heartbeat {
  int32 interval_seconds = 1;  // default 10
  int32 timeout_seconds  = 2;  // default 180 (3 minutes)
  int32 sweep_interval   = 3;  // default 10
}
```

## Heartbeat Flow

### On Heartbeat Received (read loop, FRAME_TYPE_HEARTBEAT)

1. Build HEARTBEAT echo frame, call `conn.WriteMessage`
2. Write success → `lastSuccessHeartbeat = now`, `consecutiveEchoFailures = 0`, HSET Redis
3. Write failure → `consecutiveEchoFailures++`, log + metrics counter, do NOT update lastSuccessHeartbeat

### Background Sweep (single goroutine per gateway instance)

Every `sweep_interval` seconds:

```
for each connState in localConns:
    if time.Since(state.LastSuccessHeartbeat) > timeout:
        → offline: DEL user:online:{uid}:{did}, SREM user:devices:{uid}, close conn, cm.Remove
```

### Disconnect

Existing `defer cm.Remove(uid, did)` extended with: DEL user:online key, SREM user:devices set.

## Cross-Instance Message Routing

### New gRPC: GetUserOnline

```protobuf
rpc GetUserOnline(GetUserOnlineRequest) returns (GetUserOnlineResponse);

message GetUserOnlineRequest { string user_id = 1; }
message DeviceOnlineInfo {
  string device_id = 1;
  string gateway_addr = 2;
  int64  last_heartbeat = 3;
}
message GetUserOnlineResponse { repeated DeviceOnlineInfo devices = 1; }
```

Implementation: SMEMBERS `user:devices:{userID}` → HGETALL each device key → aggregate.

### Message Service pushToReceiver Changes

1. Call any gateway's `GetUserOnline(receiverID)`
2. If devices empty → return (message already persisted, client pulls on reconnect)
3. If devices non-empty → for each device, dial `gateway_addr` directly via gRPC connection pool, call `SendToUser`
4. Gateway connection pool: `map[addr]GatewayServiceClient` with lazy init

### Gateway Address Registration

- K8s Downward API injects `GATEWAY_ADDR` = Pod IP
- Gateway `Add()` writes `GATEWAY_ADDR:9200` to Redis `user:online:{uid}:{did}`

## Offline Message Pull

### New gRPC: PullMessages

```protobuf
rpc PullMessages(PullMessagesReq) returns (PullMessagesResp);

message PullMessagesReq {
  string user_id = 1;
  int64  since_message_id = 2;
  int32  limit = 3;
}
message PullMessagesResp { repeated Message messages = 1; }
```

### Flow

1. Client sends SYNC frame with `since_message_id` after WebSocket connect
2. Gateway calls `MessageService.PullMessages(userID, sinceMsgID, limit)`
3. Message service: `SELECT * FROM private_messages WHERE receiver_id = ? AND message_server_id > ? AND is_deleted = false ORDER BY message_server_id ASC LIMIT ?`
4. Gateway pushes messages to client via WebSocket

## Files to Modify

| File | Change |
|------|--------|
| `proto/him/gateway/v1/gateway.proto` | Add GetUserOnline RPC |
| `proto/him/gateway/v1/ws.proto` | Add FRAME_TYPE_SYNC (optional) |
| `proto/him/message/v1/message.proto` | Add PullMessages RPC |
| `services/gateway/internal/conf/conf.proto` | Add Heartbeat config |
| `services/gateway/internal/biz/connmgr.go` | connState struct, sweep interface |
| `services/gateway/internal/biz/gateway.go` | Echo failure tracking, sweep goroutine |
| `services/gateway/internal/data/connmgr_redis.go` | Redis online keys, sweep implementation |
| `services/gateway/internal/data/connmgr_memory.go` | connState adapter |
| `services/gateway/internal/service/gateway.go` | Post-connect PullMessages trigger |
| `services/gateway/internal/service/gateway_grpc.go` | GetUserOnline handler |
| `services/gateway/configs/config.yaml` | Heartbeat config |
| `deploy/k8s/gateway/deployment.yaml` | GATEWAY_ADDR env |
| `services/message/internal/service/message.go` | PullMessages handler |
| `services/message/internal/biz/send.go` | pushToReceiver rewrite |
| `services/message/internal/data/gateway_client.go` | Connection pool + direct dial |
