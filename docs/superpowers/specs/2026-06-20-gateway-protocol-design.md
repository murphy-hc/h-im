# Gateway 通信协议设计

## Context

Gateway 是客户端与 IM 后端之间的唯一入口，使用 WebSocket 长连接。需要在 WebSocket 之上约定应用层帧协议，承载业务消息的序列化与路由。

## Requirements

1. **WebSocket + Protobuf**：传输层 WebSocket，载荷用 Protobuf 序列化
2. **帧头信息**：Version（1字节，协议兼容）+ FrameType（4字节，消息类别）+ PayloadLen（4字节，载荷长度）
3. **消息类别覆盖**：登录请求/响应、私聊/群聊/聊天室消息及 ACK、心跳、错误
4. **编解码封装**：在 gateway `internal/biz/` 中实现，对外暴露 `Encode` / `Decode`

## Design

### 帧格式

```
┌──────────┬────────────┬───────────────┬──────────────────┐
│ Version  │ Frame Type │ Payload Len   │ Protobuf Payload │
│ 1 byte   │ 4 bytes    │ 4 bytes       │ variable (0..N)  │
└──────────┴────────────┴───────────────┴──────────────────┘
```

- **Version**: `uint8`，当前值 `1`。协议升级时递增，用于兼容旧客户端
- **Frame Type**: `uint32`，大端序，对应 proto 中的 `FrameType` 枚举
- **Payload Len**: `uint32`，大端序，后续 payload 的字节数。心跳等无载荷消息为 `0`
- **Payload**: Protobuf 二进制，类型由 FrameType 决定

### FrameType 枚举

定义在 `proto/him/gateway/v1/` 中，与各业务 proto 共享：

| Frame Type | 枚举值 | 对应 Payload Message |
|-----------|:-----:|---------------------|
| AUTH_REQ | 1 | `AuthRequest` |
| AUTH_RESP | 2 | `AuthResponse` |
| PRIVATE_CHAT | 3 | `ChatMessage` |
| PRIVATE_ACK | 4 | `ChatAck` |
| GROUP_CHAT | 5 | `ChatMessage` |
| GROUP_ACK | 6 | `ChatAck` |
| CHATROOM_MSG | 7 | `ChatroomMessage` |
| CHATROOM_ACK | 8 | `ChatroomAck` |
| HEARTBEAT | 9 | 无 payload |
| ERROR | 10 | `ErrorMessage` |

### 编解码接口

```go
// internal/biz/codec.go
package biz

// Encode 将 frame type 和 proto message 编码为帧字节
func Encode(version uint8, frameType uint32, msg proto.Message) ([]byte, error)

// Decode 从字节中解析帧头，返回 version、frameType 和 payload 字节
func Decode(frame []byte) (version uint8, frameType uint32, payload []byte, err error)
```

### 连接流程

```
1. Client → WS Upgrade
2. Client → AUTH_REQ 帧 →
3. Client ← AUTH_RESP 帧 ← Gateway（鉴权通过返回 token/user_id）
4. 双向任意消息（PRIVATE_CHAT / GROUP_CHAT / CHATROOM_MSG ...）
5. 定时 HEARTBEAT（30s 间隔）
```

### 错误处理

- 不支持的消息版本 → ERROR 帧 + 关闭连接
- 未知 FrameType → ERROR 帧（连接保持）
- 反序列化失败 → ERROR 帧
- 心跳超时 → 主动关闭连接

### 文件结构

```
proto/him/gateway/v1/
├── ws.proto              # FrameType 枚举, 通用消息 (AuthRequest/Response, Error, Heartbeat)
├── chat.proto            # ChatMessage, ChatAck
└── chatroom.proto        # ChatroomMessage, ChatroomAck

services/gateway/internal/biz/
├── codec.go              # Encode / Decode
├── codec_test.go         # 单元测试
└── gateway.go            # HandleConnection（集成 codec + 路由）
```

### gRPC 服务（服务端推送）

Gateway 同时作为 gRPC server，供后端服务（message/group/chatroom/push 等）推送消息给在线客户端。

```protobuf
// proto/him/gateway/v1/gateway.proto
service Gateway {
  // 发送消息给指定用户（私聊）
  rpc SendToUser(SendToUserRequest) returns (SendToUserResponse);

  // 广播消息到指定群组
  rpc BroadcastToGroup(BroadcastToGroupRequest) returns (BroadcastToGroupResponse);

  // 广播消息到指定聊天室
  rpc BroadcastToChatroom(BroadcastToChatroomRequest) returns (BroadcastToChatroomResponse);

  // 发送指令给指定用户（踢出、强制下线等）
  rpc SendCommand(SendCommandRequest) returns (SendCommandResponse);
}

message SendToUserRequest {
  string user_id = 1;
  int32 frame_type = 2;   // PRIVATE_CHAT / PRIVATE_ACK / ERROR
  bytes payload = 3;       // 已序列化的 proto message
}

message BroadcastToGroupRequest {
  string group_id = 1;
  int32 frame_type = 2;
  bytes payload = 3;
  repeated string exclude_user_ids = 4;  // 可选，排除某些用户
}

message BroadcastToChatroomRequest {
  string room_id = 1;
  int32 frame_type = 2;
  int32 priority = 3;       // 聊天室消息优先级
  bytes payload = 4;
}

message SendCommandRequest {
  string user_id = 1;
  string command = 2;       // "kick", "force_logout", "kick_from_group", "kick_from_room"
  string target_id = 3;     // 群组ID或房间ID（可选）
}

message SendToUserResponse { bool success = 1; }
message BroadcastToGroupResponse { int32 delivered_count = 1; }
message BroadcastToChatroomResponse { int32 delivered_count = 1; }
message SendCommandResponse { bool success = 1; }
```

### Connection Manager (连接管理)

Gateway 维护在线用户映射，负责路由 gRPC 请求到正确的 WebSocket 连接：

```go
// internal/biz/connmgr.go
type ConnManager interface {
    // 用户管理
    Add(userID string, conn *websocket.Conn)
    Remove(userID string)
    GetConn(userID string) (*websocket.Conn, bool)

    // 群组路由
    GetGroupMembers(groupID string) []string
    JoinGroup(groupID, userID string)
    LeaveGroup(groupID, userID string)

    // 聊天室路由
    GetRoomMembers(roomID string) []string
    JoinRoom(roomID, userID string)
    LeaveRoom(roomID, userID string)

    // 统计
    OnlineCount() int
}
```

- **gRPC handler** (`internal/service/gateway_grpc.go`) 调用 `ConnManager` 获取目标连接，写入 WebSocket 帧
- 连接断开时自动从所有 group/room 中移除
- sender 自己不收到自己发的消息（`MessageService` 先推送后，gateway 已覆盖所有接收端，无需回推 sender）

### 架构更新

```
                             ┌─────────────────────────┐
Client ←── WS ──→ Gateway   │  WebSocket Handler       │
                             │  ├─ Auth (JWT verify)    │
                             │  ├─ ConnManager          │
                             │  └─ Codec (Encode/Decode)│
                   Gateway   │                          │
Backend ── gRPC ──→         │  gRPC Handler            │
(message/group/chatroom...)  │  ├─ SendToUser           │
                             │  ├─ BroadcastToGroup     │
                             │  ├─ BroadcastToChatroom  │
                             │  └─ SendCommand          │
                             └─────────────────────────┘
```

### 文件结构

```
proto/him/gateway/v1/
├── ws.proto              # FrameType 枚举, AuthReq/Resp, Error, Heartbeat
├── chat.proto            # ChatMessage, ChatAck
├── chatroom.proto        # ChatroomMessage, ChatroomAck
└── gateway.proto         # Gateway gRPC service (SendToUser, BroadcastToGroup, ...)

services/gateway/internal/
├── server/
│   ├── server.go         # WSServerProviderSet + HTTPProviderSet + GRPCProviderSet
│   ├── ws.go             # NewWSServer
│   ├── http.go           # NewHTTPServer
│   └── grpc.go           # NewGRPCServer (gateway's gRPC server)
├── service/
│   ├── service.go        # ProviderSet
│   ├── gateway.go        # GatewayService (WebSocket handler)
│   └── gateway_grpc.go   # GatewayGrpcService (gRPC handler)
├── biz/
│   ├── biz.go            # ProviderSet
│   ├── codec.go          # Encode / Decode
│   ├── connmgr.go        # ConnManager interface + in-memory impl
│   └── gateway.go        # HandleConnection
└── data/
    └── (none — no persistent state)
```

### Testing Strategy

- **Unit**: codec 编解码 round-trip 测试
- **Unit**: 边界条件（空 payload、最大 payload、无效 version/type）
- **Integration**: 真实 WebSocket 连接 + 帧收发

### Acceptance Criteria

1. Encode → Decode round-trip 无损：输入 msg 能还原
2. 短 payload（0字节）编解码正常
3. 异常输入（长度不匹配、非法 frameType）返回明确错误
4. 所有 FrameType 常量与 proto 枚举一致
