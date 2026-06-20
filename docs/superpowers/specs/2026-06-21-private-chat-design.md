# 私聊功能设计

## Context

基于 mim 项目的 message.proto 替换现有定义，实现私聊消息的发送、存储、推送、多设备同步。

## Requirements

1. **Proto 对齐 mim**：Message 结构体、MessageType、ConversationType、Attachment、ThreadReply 等全部替换
2. **不丢不重不乱序**：双 ID 机制（client_id 去重 + server_id 有序），PG 存储保证持久化
3. **多设备同步**：Gateway.SendToUser 推送所有在线设备，离线补拉
4. **按类型分表**：私聊 / 群聊 / 聊天室各自独立存储

## Design

### Proto 定义

完全替换 `proto/him/message/v1/message.proto`，引入：

- `ConversationType`: `PRIVATE=0`, `GROUP=1`, `ROOM=2`
- `MessageType`: `MSG_TYPE_TEXT=0` ~ `MSG_TYPE_CUSTOM=8`
- `SendingState`: `SENDING=0`, `SUCCEEDED=1`, `FAILED=2`
- `Message` 结构体 (19 fields)
- `Attachment` (oneof: Image/Audio/Video/File/Location)
- `ThreadReply`
- `PrivateSend` / `PrivatePush` / `PrivateAck` / `OfflineSync` / `StateSync`
- `Conversation` / `ConversationListReq` / `ConversationListResp` 等

### 消息流转

```
Client A → Gateway → Message Service
                        │
            ┌───────────┼───────────┐
            ▼           ▼           ▼
       gen server_id  INSERT PG   Gateway.SendToUser(B)
       (sequence)                  │
                                  ├─→ Client B phone
                                  └─→ Client B laptop
                                              │
                                  Client B → Gateway → Message Service
                                    PrivateAck{server_id}
```

### 有序性

`message_server_id` 由 sequence 服务分配（号段模式），严格递增。

### 去重

`message_client_id` 在 DB 中 UNIQUE 约束，重复插入跳过。

### 存储

```sql
CREATE TABLE private_messages (
    message_server_id BIGINT PRIMARY KEY,
    message_client_id VARCHAR(64) UNIQUE NOT NULL,
    sender_id         VARCHAR(64) NOT NULL,
    receiver_id       VARCHAR(64) NOT NULL,
    msg_type          INT NOT NULL DEFAULT 0,
    msg_sub_type      INT DEFAULT 0,
    text              TEXT,
    attachment        JSONB,
    server_time       BIGINT NOT NULL,
    create_time       BIGINT NOT NULL DEFAULT 0,
    is_deleted        BOOLEAN DEFAULT false,
    is_remote_read    BOOLEAN DEFAULT false
);
CREATE INDEX idx_private_sender ON private_messages(sender_id, receiver_id, message_server_id);
CREATE INDEX idx_private_receiver ON private_messages(receiver_id, message_server_id);
```

### 文件变更

| 操作 | 文件 | 职责 |
|------|------|------|
| 重写 | `proto/him/message/v1/message.proto` | 替换为 mim 结构体，增加 GROUP/ROOM 类型 |
| 修改 | `gen/go/him/message/v1/` | buf generate 重生成 |
| 新增 | `services/message/internal/biz/send.go` | SendPrivateMessage 业务逻辑 |
| 新增 | `services/message/internal/data/message_repo.go` | GORM 写入/查询/ACK |
| 新增 | `services/message/internal/data/message_model.go` | GORM Model |
| 修改 | `services/message/internal/service/message.go` | gRPC handler 对接 biz |
| 修改 | `services/message/go.mod` | 新增 sequence client 依赖 |

### Testing Strategy

- **Unit**: message_repo 写入重复 client_id 返回成功（去重）
- **Unit**: send 流程调用 sequence 生成 ID
- **Integration**: Gateway → Message → Sequence 端到端发送

### Acceptance Criteria

1. 发送消息返回 server_id
2. 重复 client_id 发送幂等
3. server_id 严格递增
4. 接收方多设备同时收到推送
5. 私聊消息独立存储在 private_messages 表
