# Private Chat Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Implement private chat with mim-compatible proto, dual-ID dedup+ordering, multi-device push.

**Architecture:** Gateway receives WS frame → calls Message Service gRPC → Message Service gets server_id from Sequence Service → persists to PG → pushes to receiver via Gateway gRPC.

**Tech Stack:** Go 1.24, Kratos v2, GORM, Protobuf, Wire

## Global Constraints

- Proto completely replaces existing `proto/him/message/v1/message.proto`
- `message_client_id` UNIQUE NOT NULL for dedup
- `message_server_id` BIGINT PRIMARY KEY from sequence service
- Private messages stored in `private_messages` table
- All gRPC clients use Kratos `transport/grpc` with `discovery:///` prefix

---

### Task 1: Replace message proto with mim-compatible definition

**Files:**
- Modify: `proto/him/message/v1/message.proto`
- Modify: `gen/go/him/message/v1/` (regenerated)

**Interfaces:**
- Produces: `Message` struct (19 fields), `ConversationType` enum (PRIVATE=0, GROUP=1, ROOM=2), `MessageType` enum (8 types), `Attachment` oneof, `ThreadReply`, `PrivateSend`, `PrivatePush`, `PrivateAck`, `OfflineSync`, `StateSync`, `Conversation`, `ConversationListReq/Resp`, `ConversationMessagesReq/Resp`, `ListConversationsResp`, `ConversationPush`

- [ ] **Step 1: Replace message.proto**

Copy mim's message.proto and augment: ConversationType adds GROUP=1, ROOM=2 (mim only has PRIVATE=0, ROOM=1). All other messages, enums, and types are identical to mim.

- [ ] **Step 2: Regenerate Go code**

Run: `~/buf/bin/buf generate`
Expected: no errors

- [ ] **Step 3: Verify**

Run: `~/buf/bin/buf lint proto/him/message/`
Expected: PASS

- [ ] **Step 4: Commit**

```bash
git add proto/him/message/v1/message.proto gen/go/him/message/v1/
git commit -m "feat(message): replace proto with mim-compatible definition"
```

---

### Task 2: Create message data layer

**Files:**
- Create: `services/message/internal/data/message_model.go`
- Create: `services/message/internal/data/message_repo.go`
- Modify: `services/message/internal/data/data.go`

**Interfaces:**
- Consumes: `*gorm.DB` from `Data`
- Produces: `MessageModel` struct, `MessageRepo` with `Insert(ctx, msg *MessageModel) error`, `FindByServerID(ctx, id int64) (*MessageModel, error)`, `MarkRead(ctx, serverID int64, userID string) error`

- [ ] **Step 1: Create MessageModel**

```go
type MessageModel struct {
    MessageServerID int64  `gorm:"column:message_server_id;primaryKey"`
    MessageClientID string `gorm:"column:message_client_id;uniqueIndex;size:64;not null"`
    SenderID        string `gorm:"column:sender_id;size:64;not null;index:idx_sender"`
    ReceiverID      string `gorm:"column:receiver_id;size:64;not null;index:idx_receiver"`
    ConvType        int32  `gorm:"column:conv_type;not null"`
    MsgType         int32  `gorm:"column:msg_type;not null;default:0"`
    MsgSubType      int32  `gorm:"column:msg_sub_type;default:0"`
    Text            string `gorm:"column:text"`
    Attachment      string `gorm:"column:attachment;type:jsonb"`
    ServerTime      int64  `gorm:"column:server_time;not null"`
    CreateTime      int64  `gorm:"column:create_time;not null;default:0"`
    IsDeleted       bool   `gorm:"column:is_deleted;default:false"`
    IsRemoteRead    bool   `gorm:"column:is_remote_read;default:false"`
}
func (MessageModel) TableName() string { return "private_messages" }
```

- [ ] **Step 2: Create MessageRepo**

```go
type MessageRepo struct { db *gorm.DB }
func NewMessageRepo(data *Data) *MessageRepo { return &MessageRepo{db: data.DB} }

func (r *MessageRepo) Insert(ctx context.Context, m *MessageModel) error {
    return r.db.WithContext(ctx).Create(m).Error
}

func (r *MessageRepo) MarkRemoteRead(ctx context.Context, serverID int64) error {
    return r.db.WithContext(ctx).Model(&MessageModel{}).Where("message_server_id = ?", serverID).Update("is_remote_read", true).Error
}
```

- [ ] **Step 3: Update data.go ProviderSet**

Add `NewMessageRepo` to ProviderSet.

- [ ] **Step 4: Add AutoMigrate**

In `Migrate()` method, add `d.DB.AutoMigrate(&MessageModel{})`.

- [ ] **Step 5: Commit**

```bash
git add services/message/internal/data/
git commit -m "feat(message): add private message data layer"
```

---

### Task 3: Implement send message biz logic

**Files:**
- Create: `services/message/internal/biz/send.go`
- Modify: `services/message/internal/biz/biz.go`

**Interfaces:**
- Consumes: `MessageRepo.Insert`, `SequenceServiceClient.NextBatchID`
- Produces: `SendUseCase.SendPrivateMessage(ctx, req) (serverID int64, err error)`

- [ ] **Step 1: Create send.go**

```go
type SendUseCase struct {
    repo      *data.MessageRepo
    seqClient pb.SequenceServiceClient
}

func NewSendUseCase(repo *data.MessageRepo, seqClient pb.SequenceServiceClient) *SendUseCase {
    return &SendUseCase{repo: repo, seqClient: seqClient}
}

func (uc *SendUseCase) SendPrivateMessage(ctx context.Context, senderID, receiverID string, msgType int32, text string, clientID string) (int64, error) {
    resp, err := uc.seqClient.NextBatchID(ctx, &pb.NextBatchIDRequest{Key: "private_message_id", Size: 1})
    if err != nil { return 0, err }
    serverID := resp.Start
    now := time.Now().UnixMilli()
    m := &data.MessageModel{
        MessageServerID: serverID,
        MessageClientID: clientID,
        SenderID:        senderID,
        ReceiverID:      receiverID,
        ConvType:        int32(messagev1.ConversationType_CONVERSATION_PRIVATE),
        MsgType:         msgType,
        Text:            text,
        ServerTime:      now,
        CreateTime:      now,
    }
    if err := uc.repo.Insert(ctx, m); err != nil {
        return 0, err
    }
    return serverID, nil
}
```

- [ ] **Step 2: Update biz.go ProviderSet**

Add `NewSendUseCase` to ProviderSet.

- [ ] **Step 3: Add sequence client to wire**

```go
// cmd/server/wire.go
func NewSequenceClient(conn *grpc.ClientConn) pb.SequenceServiceClient {
    return pb.NewSequenceServiceClient(conn)
}
```

- [ ] **Step 4: Wire the gRPC client connection**

Create `services/message/internal/server/grpc_client.go` that creates a Kratos gRPC client dialing sequence:
```go
func NewSequenceGRPCClient() (*grpc.ClientConn, func(), error) {
    conn, err := grpc.DialInsecure(context.Background(),
        grpc.WithEndpoint("discovery:///sequence.default.svc.cluster.local:9108"),
    )
    if err != nil { return nil, nil, err }
    return conn, func() { conn.Close() }, nil
}
```

- [ ] **Step 5: Verify build**

Run: `cd services/message && wire ./cmd/server/ && go build ./cmd/server/`
Expected: PASS

- [ ] **Step 6: Commit**

```bash
git add services/message/
git commit -m "feat(message): implement private message send logic with sequence client"
```

---

### Task 4: Update service layer and final verification

**Files:**
- Modify: `services/message/internal/service/message.go`
- Modify: `services/message/internal/service/service.go`

- [ ] **Step 1: Implement SendMessage gRPC handler**

```go
func (s *MessageService) SendMessage(ctx context.Context, req *pb.SendMessageRequest) (*pb.SendMessageResponse, error) {
    serverID, err := s.sendUC.SendPrivateMessage(ctx, req.SenderId, req.ReceiverId, int32(req.MsgType), req.Text, req.MessageClientId)
    if err != nil { return nil, err }
    return &pb.SendMessageResponse{MessageServerId: serverID, ServerTime: time.Now().UnixMilli()}, nil
}
```

- [ ] **Step 2: Full verification**

Run: `cd services/message && go vet ./... && go test -short ./...`
Expected: PASS

- [ ] **Step 3: Build all services**

Run: `make build`
Expected: all 10 services PASS

- [ ] **Step 4: Commit**

```bash
git add -A
git commit -m "feat(message): complete private message send flow"
```
