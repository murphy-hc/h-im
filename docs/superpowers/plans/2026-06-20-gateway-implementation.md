# Gateway Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task.

**Goal:** 实现 gateway 服务：WebSocket 帧协议编解码、in-memory 连接管理、gRPC push 接口、WS handler。

**Architecture:** 客户端通过 WebSocket 连接 gateway，gateway 通过 gRPC 接收后端推送，ConnManager 维护在线用户/群组/聊天室路由。

**Tech Stack:** Go 1.24, Kratos v2, gorilla/websocket, Protobuf, Wire

## Global Constraints

- Version 固定 `uint8(1)`
- FrameType 用 `uint32` 大端序
- PayloadLen 用 `uint32` 大端序
- 所有 proto 放在 `proto/him/gateway/v1/`
- ConnManager 为 in-memory 实现，不需要持久化
- Gateway gRPC server 端口：`:9200`
- Gateway WS 端口：`:8080`，HTTP metrics 端口：`:8180`

---

### Task 1: Define gateway proto files

**Files:**
- Create: `proto/him/gateway/v1/ws.proto`
- Create: `proto/him/gateway/v1/chat.proto`
- Create: `proto/him/gateway/v1/chatroom.proto`
- Create: `proto/him/gateway/v1/gateway.proto`
- Modify: `buf.gen.yaml` (add gateway proto path)

**Interfaces:**
- Produces: `FrameType` enum, `AuthRequest/Response`, `ErrorMessage`, `ChatMessage`, `ChatAck`, `ChatroomMessage`, `ChatroomAck`, `Gateway` gRPC service

- [ ] **Step 1: Create ws.proto**

```protobuf
syntax = "proto3";
package him.gateway.v1;
option go_package = "github.com/murphy-hc/h-im/gen/go/him/gateway/v1;gatewayv1";

enum FrameType {
  FRAME_TYPE_UNSPECIFIED  = 0;
  FRAME_TYPE_AUTH_REQ     = 1;
  FRAME_TYPE_AUTH_RESP    = 2;
  FRAME_TYPE_PRIVATE_CHAT = 3;
  FRAME_TYPE_PRIVATE_ACK  = 4;
  FRAME_TYPE_GROUP_CHAT   = 5;
  FRAME_TYPE_GROUP_ACK    = 6;
  FRAME_TYPE_CHATROOM_MSG = 7;
  FRAME_TYPE_CHATROOM_ACK = 8;
  FRAME_TYPE_HEARTBEAT    = 9;
  FRAME_TYPE_ERROR        = 10;
}

message AuthRequest { string token = 1; }
message AuthResponse { string user_id = 1; bool success = 2; string error = 3; }
message ErrorMessage { int32 code = 1; string message = 2; }
```

- [ ] **Step 2: Create chat.proto**

```protobuf
syntax = "proto3";
package him.gateway.v1;
option go_package = "github.com/murphy-hc/h-im/gen/go/him/gateway/v1;gatewayv1";

message ChatMessage {
  string message_id = 1;
  string sender_id  = 2;
  string content    = 3;
  int64  timestamp  = 4;
  string reply_to   = 5;
  int32  msg_type   = 6;  // text/image/voice/video/file
}

message ChatAck {
  string message_id = 1;
  string user_id    = 2;
  int64  timestamp  = 3;
}
```

- [ ] **Step 3: Create chatroom.proto**

```protobuf
syntax = "proto3";
package him.gateway.v1;
option go_package = "github.com/murphy-hc/h-im/gen/go/him/gateway/v1;gatewayv1";

message ChatroomMessage {
  string message_id = 1;
  string room_id    = 2;
  string sender_id  = 3;
  string content    = 4;
  int64  timestamp  = 5;
  int32  priority   = 6;
}

message ChatroomAck {
  string message_id = 1;
  string user_id    = 2;
  int64  timestamp  = 3;
}
```

- [ ] **Step 4: Create gateway.proto**

```protobuf
syntax = "proto3";
package him.gateway.v1;
option go_package = "github.com/murphy-hc/h-im/gen/go/him/gateway/v1;gatewayv1";

service Gateway {
  rpc SendToUser(SendToUserRequest)           returns (SendToUserResponse);
  rpc BroadcastToGroup(BroadcastToGroupRequest) returns (BroadcastToGroupResponse);
  rpc BroadcastToChatroom(BroadcastToChatroomRequest) returns (BroadcastToChatroomResponse);
  rpc SendCommand(SendCommandRequest)         returns (SendCommandResponse);
}

message SendToUserRequest {
  string user_id   = 1;
  int32  frame_type = 2;
  bytes  payload   = 3;
}
message SendToUserResponse { bool success = 1; }

message BroadcastToGroupRequest {
  string group_id          = 1;
  int32  frame_type         = 2;
  bytes  payload           = 3;
  repeated string exclude_user_ids = 4;
}
message BroadcastToGroupResponse { int32 delivered_count = 1; }

message BroadcastToChatroomRequest {
  string room_id   = 1;
  int32  frame_type = 2;
  int32  priority  = 3;
  bytes  payload   = 4;
}
message BroadcastToChatroomResponse { int32 delivered_count = 1; }

message SendCommandRequest {
  string user_id   = 1;
  string command   = 2;
  string target_id = 3;
}
message SendCommandResponse { bool success = 1; }
```

- [ ] **Step 5: Regenerate Go code**

Run: `~/buf/bin/buf generate`
Expected: no errors, `gen/go/him/gateway/v1/` populated

- [ ] **Step 6: Commit**

```bash
git add proto/him/gateway/ gen/go/him/gateway/
git commit -m "feat(gateway): add proto definitions for WS frames and gRPC push"
```

---

### Task 2: Implement frame codec

**Files:**
- Create: `services/gateway/internal/biz/codec.go`
- Create: `services/gateway/internal/biz/codec_test.go`

**Interfaces:**
- Produces: `Encode(version uint8, frameType gatewayv1.FrameType, msg proto.Message) ([]byte, error)`
- Produces: `Decode(frame []byte) (version uint8, frameType gatewayv1.FrameType, payload []byte, err error)`

- [ ] **Step 1: Write failing test**

Create `services/gateway/internal/biz/codec_test.go`:

```go
package biz_test

import (
	"testing"
	gatewayv1 "github.com/murphy-hc/h-im/gen/go/him/gateway/v1"
	"github.com/murphy-hc/h-im/services/gateway/internal/biz"
	"google.golang.org/protobuf/proto"
)

func TestEncodeDecodeRoundTrip(t *testing.T) {
	original := &gatewayv1.AuthRequest{Token: "test-token"}
	frame, err := biz.Encode(1, gatewayv1.FrameType_FRAME_TYPE_AUTH_REQ, original)
	if err != nil { t.Fatalf("encode: %v", err) }
	version, ft, payload, err := biz.Decode(frame)
	if err != nil { t.Fatalf("decode: %v", err) }
	if version != 1 { t.Errorf("version = %d, want 1", version) }
	if ft != gatewayv1.FrameType_FRAME_TYPE_AUTH_REQ { t.Errorf("wrong frame type") }
	var decoded gatewayv1.AuthRequest
	if err := proto.Unmarshal(payload, &decoded); err != nil { t.Fatalf("unmarshal: %v", err) }
	if decoded.Token != original.Token { t.Errorf("token mismatch") }
}

func TestDecodeInvalidFrame(t *testing.T) {
	_, _, _, err := biz.Decode([]byte{})
	if err == nil { t.Fatal("expected error for empty frame") }
}

func TestEncodeEmptyPayload(t *testing.T) {
	frame, err := biz.Encode(1, gatewayv1.FrameType_FRAME_TYPE_HEARTBEAT, nil)
	if err != nil { t.Fatalf("encode: %v", err) }
	if len(frame) != 9 { t.Errorf("expected 9 bytes (1+4+4+0), got %d", len(frame)) }
	_, ft, payload, _ := biz.Decode(frame)
	if ft != gatewayv1.FrameType_FRAME_TYPE_HEARTBEAT { t.Errorf("wrong frame type") }
	if len(payload) != 0 { t.Errorf("expected empty payload") }
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `cd services/gateway && go test ./internal/biz/... -v -run TestEncode`
Expected: FAIL — "undefined: biz.Encode"

- [ ] **Step 3: Implement codec**

Create `services/gateway/internal/biz/codec.go`:

```go
package biz

import (
	"encoding/binary"
	"fmt"
	gatewayv1 "github.com/murphy-hc/h-im/gen/go/him/gateway/v1"
	"google.golang.org/protobuf/proto"
)

const HeaderSize = 9
const CurrentVersion uint8 = 1

func Encode(version uint8, frameType gatewayv1.FrameType, msg proto.Message) ([]byte, error) {
	var payload []byte
	if msg != nil {
		var err error
		payload, err = proto.Marshal(msg)
		if err != nil { return nil, fmt.Errorf("marshal: %w", err) }
	}
	buf := make([]byte, HeaderSize+len(payload))
	buf[0] = version
	binary.BigEndian.PutUint32(buf[1:5], uint32(frameType))
	binary.BigEndian.PutUint32(buf[5:9], uint32(len(payload)))
	copy(buf[9:], payload)
	return buf, nil
}

func Decode(frame []byte) (version uint8, frameType gatewayv1.FrameType, payload []byte, err error) {
	if len(frame) < HeaderSize { return 0, 0, nil, fmt.Errorf("frame too short: %d < %d", len(frame), HeaderSize) }
	version = frame[0]
	if version != CurrentVersion { return version, 0, nil, fmt.Errorf("unsupported version: %d", version) }
	frameType = gatewayv1.FrameType(binary.BigEndian.Uint32(frame[1:5]))
	payloadLen := binary.BigEndian.Uint32(frame[5:9])
	if uint32(len(frame)) < HeaderSize+payloadLen { return 0, 0, nil, fmt.Errorf("frame length mismatch") }
	payload = frame[9 : 9+payloadLen]
	return version, frameType, payload, nil
}
```

- [ ] **Step 4: Run tests**

Run: `cd services/gateway && go test ./internal/biz/... -v -run TestEncode`
Expected: 3 tests PASS

- [ ] **Step 5: Commit**

```bash
git add services/gateway/internal/biz/
git commit -m "feat(gateway): implement WS frame codec with round-trip tests"
```

---

### Task 3: Implement ConnManager

**Files:**
- Create: `services/gateway/internal/biz/connmgr.go`
- Create: `services/gateway/internal/biz/connmgr_test.go`
- Modify: `services/gateway/internal/biz/biz.go` (add to ProviderSet)

**Interfaces:**
- Consumes: `gorilla/websocket.Conn`
- Produces: `ConnManager` interface + `connManager` struct in ProviderSet

- [ ] **Step 1: Write failing test**

Create `services/gateway/internal/biz/connmgr_test.go`:

```go
package biz_test

import (
	"testing"
	"github.com/murphy-hc/h-im/services/gateway/internal/biz"
)

func TestConnManagerAddRemove(t *testing.T) {
	cm := biz.NewConnManager()
	if c := cm.OnlineCount(); c != 0 { t.Errorf("want 0, got %d", c) }
	cm.Add("u1", nil)
	if c := cm.OnlineCount(); c != 1 { t.Errorf("want 1, got %d", c) }
	_, ok := cm.GetConn("u1")
	if !ok { t.Error("u1 not found") }
	cm.Remove("u1")
	if c := cm.OnlineCount(); c != 0 { t.Errorf("want 0, got %d", c) }
}

func TestConnManagerGroupRouting(t *testing.T) {
	cm := biz.NewConnManager()
	cm.Add("u1", nil)
	cm.Add("u2", nil)
	cm.Add("u3", nil)
	cm.JoinGroup("g1", "u1")
	cm.JoinGroup("g1", "u2")
	cm.JoinGroup("g1", "u3")
	members := cm.GetGroupMembers("g1")
	if len(members) != 3 { t.Errorf("want 3, got %d", len(members)) }
	cm.LeaveGroup("g1", "u2")
	if len(cm.GetGroupMembers("g1")) != 2 { t.Error("u2 not removed") }
	cm.Remove("u3")
	if len(cm.GetGroupMembers("g1")) != 1 { t.Error("u3 not cleaned up") }
}

func TestConnManagerRoomRouting(t *testing.T) {
	cm := biz.NewConnManager()
	cm.Add("u1", nil)
	cm.Add("u2", nil)
	cm.JoinRoom("r1", "u1")
	cm.JoinRoom("r1", "u2")
	if len(cm.GetRoomMembers("r1")) != 2 { t.Errorf("want 2, got %d", len(cm.GetRoomMembers("r1"))) }
	cm.LeaveRoom("r1", "u2")
	if len(cm.GetRoomMembers("r1")) != 1 { t.Error("u2 not removed") }
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `cd services/gateway && go test ./internal/biz/... -v -run TestConnManager`
Expected: FAIL — "undefined: biz.NewConnManager"

- [ ] **Step 3: Implement ConnManager**

Create `services/gateway/internal/biz/connmgr.go`:

```go
package biz

import (
	"sync"
	"github.com/gorilla/websocket"
)

type ConnManager interface {
	Add(userID string, conn *websocket.Conn)
	Remove(userID string)
	GetConn(userID string) (*websocket.Conn, bool)
	GetGroupMembers(groupID string) []string
	JoinGroup(groupID, userID string)
	LeaveGroup(groupID, userID string)
	GetRoomMembers(roomID string) []string
	JoinRoom(roomID, userID string)
	LeaveRoom(roomID, userID string)
	OnlineCount() int
}

type connManager struct {
	mu          sync.RWMutex
	conns       map[string]*websocket.Conn
	groupUsers  map[string]map[string]struct{}
	roomUsers   map[string]map[string]struct{}
	userGroups  map[string]map[string]struct{}
	userRooms   map[string]map[string]struct{}
}

func NewConnManager() ConnManager {
	return &connManager{
		conns:      make(map[string]*websocket.Conn),
		groupUsers: make(map[string]map[string]struct{}),
		roomUsers:  make(map[string]map[string]struct{}),
		userGroups: make(map[string]map[string]struct{}),
		userRooms:  make(map[string]map[string]struct{}),
	}
}

func (cm *connManager) Add(userID string, conn *websocket.Conn) {
	cm.mu.Lock(); defer cm.mu.Unlock()
	cm.conns[userID] = conn
}

func (cm *connManager) Remove(userID string) {
	cm.mu.Lock(); defer cm.mu.Unlock()
	delete(cm.conns, userID)
	for gid := range cm.userGroups[userID] {
		delete(cm.groupUsers[gid], userID)
	}
	for rid := range cm.userRooms[userID] {
		delete(cm.roomUsers[rid], userID)
	}
	delete(cm.userGroups, userID)
	delete(cm.userRooms, userID)
}

func (cm *connManager) GetConn(userID string) (*websocket.Conn, bool) {
	cm.mu.RLock(); defer cm.mu.RUnlock()
	c, ok := cm.conns[userID]
	return c, ok
}

func (cm *connManager) JoinGroup(groupID, userID string) {
	cm.mu.Lock(); defer cm.mu.Unlock()
	if cm.groupUsers[groupID] == nil { cm.groupUsers[groupID] = make(map[string]struct{}) }
	cm.groupUsers[groupID][userID] = struct{}{}
	if cm.userGroups[userID] == nil { cm.userGroups[userID] = make(map[string]struct{}) }
	cm.userGroups[userID][groupID] = struct{}{}
}

func (cm *connManager) LeaveGroup(groupID, userID string) {
	cm.mu.Lock(); defer cm.mu.Unlock()
	delete(cm.groupUsers[groupID], userID)
	delete(cm.userGroups[userID], groupID)
}

func (cm *connManager) GetGroupMembers(groupID string) []string {
	cm.mu.RLock(); defer cm.mu.RUnlock()
	var ids []string
	for id := range cm.groupUsers[groupID] { ids = append(ids, id) }
	return ids
}

// JoinRoom / LeaveRoom / GetRoomMembers same pattern
func (cm *connManager) JoinRoom(roomID, userID string) {
	cm.mu.Lock(); defer cm.mu.Unlock()
	if cm.roomUsers[roomID] == nil { cm.roomUsers[roomID] = make(map[string]struct{}) }
	cm.roomUsers[roomID][userID] = struct{}{}
	if cm.userRooms[userID] == nil { cm.userRooms[userID] = make(map[string]struct{}) }
	cm.userRooms[userID][roomID] = struct{}{}
}
func (cm *connManager) LeaveRoom(roomID, userID string) {
	cm.mu.Lock(); defer cm.mu.Unlock()
	delete(cm.roomUsers[roomID], userID)
	delete(cm.userRooms[userID], roomID)
}
func (cm *connManager) GetRoomMembers(roomID string) []string {
	cm.mu.RLock(); defer cm.mu.RUnlock()
	var ids []string
	for id := range cm.roomUsers[roomID] { ids = append(ids, id) }
	return ids
}
func (cm *connManager) OnlineCount() int {
	cm.mu.RLock(); defer cm.mu.RUnlock()
	return len(cm.conns)
}
```

- [ ] **Step 4: Update biz.go ProviderSet**

Replace `services/gateway/internal/biz/biz.go`:

```go
package biz
import "github.com/google/wire"
var ProviderSet = wire.NewSet(NewGatewayUseCase, NewConnManager)
```

- [ ] **Step 5: Run tests**

Run: `cd services/gateway && go test ./internal/biz/... -v -run TestConnManager`
Expected: 3 tests PASS

- [ ] **Step 6: Commit**

```bash
git add services/gateway/internal/biz/
git commit -m "feat(gateway): implement in-memory ConnManager with tests"
```

---

### Task 4: Implement Gateway gRPC service

**Files:**
- Create: `services/gateway/internal/service/gateway_grpc.go`
- Modify: `services/gateway/internal/service/service.go` (add to ProviderSet)
- Modify: `services/gateway/internal/server/server.go` (add GRPCProviderSet)
- Create: `services/gateway/internal/server/grpc.go`

**Interfaces:**
- Consumes: `ConnManager`, `biz.Encode`
- Produces: `GatewayGrpcService` implementing `gatewayv1.GatewayServer`

- [ ] **Step 1: Create gRPC service**

Create `services/gateway/internal/service/gateway_grpc.go`:

```go
package service

import (
	"context"
	"fmt"

	gatewayv1 "github.com/murphy-hc/h-im/gen/go/him/gateway/v1"
	"github.com/murphy-hc/h-im/services/gateway/internal/biz"
)

type GatewayGrpcService struct {
	gatewayv1.UnimplementedGatewayServer
	cm *biz.ConnManager
}

func NewGatewayGrpcService(cm *biz.ConnManager) *GatewayGrpcService {
	return &GatewayGrpcService{cm: cm}
}

func (s *GatewayGrpcService) SendToUser(ctx context.Context, req *gatewayv1.SendToUserRequest) (*gatewayv1.SendToUserResponse, error) {
	conn, ok := (*s.cm).GetConn(req.UserId)
	if !ok { return &gatewayv1.SendToUserResponse{Success: false}, nil }
	ft := gatewayv1.FrameType(req.FrameType)
	frame, err := biz.Encode(biz.CurrentVersion, ft, nil)
	if err != nil { return nil, fmt.Errorf("encode: %w", err) }
	// Write payload directly
	msg := make([]byte, len(frame)+len(req.Payload))
	copy(msg, frame)
	copy(msg[len(frame):], req.Payload)
	if err := conn.WriteMessage(websocket.BinaryMessage, msg); err != nil {
		return &gatewayv1.SendToUserResponse{Success: false}, nil
	}
	return &gatewayv1.SendToUserResponse{Success: true}, nil
}

func (s *GatewayGrpcService) BroadcastToGroup(ctx context.Context, req *gatewayv1.BroadcastToGroupRequest) (*gatewayv1.BroadcastToGroupResponse, error) {
	members := (*s.cm).GetGroupMembers(req.GroupId)
	ft := gatewayv1.FrameType(req.FrameType)
	frame, _ := biz.Encode(biz.CurrentVersion, ft, nil)
	msg := make([]byte, len(frame)+len(req.Payload))
	copy(msg, frame)
	copy(msg[len(frame):], req.Payload)
	exclude := make(map[string]bool)
	for _, uid := range req.ExcludeUserIds { exclude[uid] = true }
	var delivered int32
	for _, uid := range members {
		if exclude[uid] { continue }
		conn, ok := (*s.cm).GetConn(uid)
		if ok {
			conn.WriteMessage(websocket.BinaryMessage, msg)
			delivered++
		}
	}
	return &gatewayv1.BroadcastToGroupResponse{DeliveredCount: delivered}, nil
}

func (s *GatewayGrpcService) BroadcastToChatroom(ctx context.Context, req *gatewayv1.BroadcastToChatroomRequest) (*gatewayv1.BroadcastToChatroomResponse, error) {
	members := (*s.cm).GetRoomMembers(req.RoomId)
	ft := gatewayv1.FrameType(req.FrameType)
	frame, _ := biz.Encode(biz.CurrentVersion, ft, nil)
	msg := make([]byte, len(frame)+len(req.Payload))
	copy(msg, frame)
	copy(msg[len(frame):], req.Payload)
	var delivered int32
	for _, uid := range members {
		if conn, ok := (*s.cm).GetConn(uid); ok {
			conn.WriteMessage(websocket.BinaryMessage, msg)
			delivered++
		}
	}
	return &gatewayv1.BroadcastToChatroomResponse{DeliveredCount: delivered}, nil
}

func (s *GatewayGrpcService) SendCommand(ctx context.Context, req *gatewayv1.SendCommandRequest) (*gatewayv1.SendCommandResponse, error) {
	conn, ok := (*s.cm).GetConn(req.UserId)
	if !ok { return &gatewayv1.SendCommandResponse{Success: false}, nil }
	frame, _ := biz.Encode(biz.CurrentVersion, gatewayv1.FrameType_FRAME_TYPE_ERROR, nil)
	conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, req.Command))
	_ = frame
	_ = conn
	return &gatewayv1.SendCommandResponse{Success: true}, nil
}
```

- [ ] **Step 2: Update service.go ProviderSet**

Replace `services/gateway/internal/service/service.go`:

```go
package service
import "github.com/google/wire"
var ProviderSet = wire.NewSet(NewGatewayService, NewGatewayGrpcService)
```

- [ ] **Step 3: Create server/grpc.go**

Create `services/gateway/internal/server/grpc.go`:

```go
package server
import (
	gatewayv1 "github.com/murphy-hc/h-im/gen/go/him/gateway/v1"
	"github.com/go-kratos/kratos/v2/middleware/metrics"
	"github.com/go-kratos/kratos/v2/middleware/recovery"
	"github.com/go-kratos/kratos/v2/middleware/tracing"
	"github.com/go-kratos/kratos/v2/transport/grpc"
	"go.opentelemetry.io/otel/metric"
	"github.com/murphy-hc/h-im/services/gateway/internal/conf"
)
func NewGRPCServer(bc *conf.Bootstrap, meter metric.Meter, svc *service.GatewayGrpcService) *grpc.Server {
	counter, _ := metrics.DefaultRequestsCounter(meter, metrics.DefaultServerRequestsCounterName)
	histogram, _ := metrics.DefaultSecondsHistogram(meter, metrics.DefaultServerSecondsHistogramName)
	srv := grpc.NewServer(
		grpc.Address(":9200"),
		grpc.Middleware(recovery.Recovery(), tracing.Server(), metrics.Server(metrics.WithRequests(counter), metrics.WithSeconds(histogram))),
	)
	gatewayv1.RegisterGatewayServer(srv, svc)
	return srv
}
```

- [ ] **Step 4: Update server.go**

Replace `services/gateway/internal/server/server.go`:

```go
package server
import (
	"github.com/google/wire"
	"github.com/murphy-hc/h-im/services/gateway/internal/conf"
)
var WSServerProviderSet  = wire.NewSet(NewWSServer, wire.FieldsOf(new(*conf.Bootstrap), "Server"))
var HTTPProviderSet      = wire.NewSet(NewHTTPServer)
var GRPCProviderSet      = wire.NewSet(NewGRPCServer)
```

- [ ] **Step 5: Update wire.go**

Replace `services/gateway/cmd/server/wire.go`:

```go
//go:build wireinject
// +build wireinject
package main
import (
	"github.com/go-kratos/kratos/v2"
	"github.com/google/wire"
	"go.opentelemetry.io/otel/metric"
	"github.com/murphy-hc/h-im/services/gateway/internal/biz"
	"github.com/murphy-hc/h-im/services/gateway/internal/conf"
	"github.com/murphy-hc/h-im/services/gateway/internal/server"
	"github.com/murphy-hc/h-im/services/gateway/internal/service"
)
func wireApp(bc *conf.Bootstrap, meter metric.Meter) (*kratos.App, func(), error) {
	panic(wire.Build(
		server.WSServerProviderSet,
		server.HTTPProviderSet,
		server.GRPCProviderSet,
		service.ProviderSet,
		biz.ProviderSet,
		newApp,
	))
}
```

- [ ] **Step 6: Regenerate wire and build**

Run:
```bash
cd services/gateway && wire ./cmd/server/ && go build ./cmd/server/
```
Expected: wire writes wire_gen.go, build passes

- [ ] **Step 7: Commit**

```bash
git add services/gateway/
git commit -m "feat(gateway): implement gRPC push service and wire integration"
```

---

### Task 5: Rewrite WS handler with ConnManager + Codec

**Files:**
- Modify: `services/gateway/internal/biz/gateway.go`
- Modify: `services/gateway/internal/service/gateway.go`
- Modify: `services/gateway/internal/server/ws.go`

- [ ] **Step 1: Rewrite biz/gateway.go**

```go
package biz
import (
	"context"
	"fmt"
	gatewayv1 "github.com/murphy-hc/h-im/gen/go/him/gateway/v1"
	"github.com/gorilla/websocket"
	"github.com/go-kratos/kratos/v2/log"
	"google.golang.org/protobuf/proto"
	"time"
)
type GatewayUseCase struct {
	cm        ConnManager
	authClients ... // will be injected later
}
func NewGatewayUseCase(cm ConnManager) *GatewayUseCase { return &GatewayUseCase{cm: cm} }
func (uc *GatewayUseCase) HandleConnection(ctx context.Context, conn *websocket.Conn, userID string) {
	defer uc.cm.Remove(userID)
	conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	conn.SetPongHandler(func(string) error {
		conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})
	for {
		_, raw, err := conn.ReadMessage()
		if err != nil { break }
		version, ft, payload, err := Decode(raw)
		if err != nil {
			frame, _ := Encode(CurrentVersion, gatewayv1.FrameType_FRAME_TYPE_ERROR,
				&gatewayv1.ErrorMessage{Code: 1, Message: err.Error()})
			conn.WriteMessage(websocket.BinaryMessage, frame)
			continue
		}
		_ = version
		switch ft {
		case gatewayv1.FrameType_FRAME_TYPE_HEARTBEAT:
			// respond with heartbeat echo
			frame, _ := Encode(CurrentVersion, gatewayv1.FrameType_FRAME_TYPE_HEARTBEAT, nil)
			conn.WriteMessage(websocket.BinaryMessage, frame)
		default:
			// For client-to-server messages, route to backend via gRPC (future tasks)
			_ = payload
			_ = ft
		}
	}
}
```

- [ ] **Step 2: Update service/gateway.go** — use ConnManager

```go
package service
import (
	"log"
	"net/http"
	"github.com/gorilla/websocket"
	"github.com/murphy-hc/h-im/services/gateway/internal/biz"
)
var upgrader = websocket.Upgrader{ReadBufferSize: 1024, WriteBufferSize: 1024}
type GatewayService struct { uc *biz.GatewayUseCase; cm *biz.ConnManager }
func NewGatewayService(uc *biz.GatewayUseCase, cm *biz.ConnManager) *GatewayService {
	return &GatewayService{uc: uc, cm: cm}
}
func (s *GatewayService) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil { log.Printf("ws upgrade: %v", err); return }
	defer conn.Close()
	// For now, generate anonymous ID (auth to be added later)
	userID := r.URL.Query().Get("user_id")
	if userID == "" { userID = "anon-" + r.RemoteAddr }
	s.cm.Add(userID, conn)
	s.uc.HandleConnection(r.Context(), conn, userID)
}
```

- [ ] **Step 3: Build**

Run: `cd services/gateway && wire ./cmd/server/ && go build ./cmd/server/`
Expected: build passes

- [ ] **Step 4: Commit**

```bash
git add services/gateway/
git commit -m "feat(gateway): integrate ConnManager and codec into WS handler"
```

---

### Task 6: Final verification

- [ ] **Step 1: Build**

Run: `cd services/gateway && go build ./cmd/server/`
Expected: PASS

- [ ] **Step 2: Vet**

Run: `cd services/gateway && go vet ./...`
Expected: PASS

- [ ] **Step 3: Tests**

Run: `cd services/gateway && go test -short -count=1 ./...`
Expected: codec tests + connmgr tests PASS

- [ ] **Step 4: buf lint**

Run: `~/buf/bin/buf lint`
Expected: PASS

- [ ] **Step 5: Clean up and commit**

```bash
rm -f bin/server
git add -A && git diff --cached --stat
git commit -m "chore(gateway): final verification and cleanup"
```
