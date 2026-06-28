package biz

import (
	"context"
	"sync"
	"testing"

	msgpb "github.com/murphy-hc/h-im/gen/go/him/message/v1"
	pb "github.com/murphy-hc/h-im/gen/go/him/sequence/v1"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/proto"
)

// mockMessageRepo implements MessageRepo for testing.
type mockMessageRepo struct {
	mu       sync.Mutex
	messages []*Message
	inserts  int
}

func (m *mockMessageRepo) Insert(ctx context.Context, msg *Message) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.messages = append(m.messages, msg)
	m.inserts++
	return nil
}
func (m *mockMessageRepo) InsertChatroom(ctx context.Context, serverID int64, clientID, roomID, senderID string, msgType int32, text, attachment string, serverTime int64) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.inserts++
	return nil
}
func (m *mockMessageRepo) InsertGroup(ctx context.Context, serverID int64, clientID, groupID, senderID string, msgType int32, text, attachment string, serverTime int64) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.inserts++
	return nil
}
func (m *mockMessageRepo) GetReceiverID(ctx context.Context, serverID int64) (string, error) {
	return "receiver-1", nil
}
func (m *mockMessageRepo) MarkDelivered(ctx context.Context, serverID int64) error { return nil }
func (m *mockMessageRepo) MarkRead(ctx context.Context, serverID int64) error      { return nil }
func (m *mockMessageRepo) MarkRecalled(ctx context.Context, serverID int64, senderID string, serverTime int64) (bool, error) {
	return true, nil
}
func (m *mockMessageRepo) PullSince(ctx context.Context, userID string, sinceID int64, limit int32) ([]Message, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	out := make([]Message, len(m.messages))
	for i, msg := range m.messages {
		out[i] = *msg
	}
	return out, nil
}

// mockMessageGateway implements MessageGateway for testing.
type mockMessageGateway struct {
	sentToDevice    []string
	broadcastRoom   []string
	broadcastGroup  []string
}

func (g *mockMessageGateway) SendToDevice(ctx context.Context, gatewayAddr, userID string, frameType int32, payload []byte) error {
	g.sentToDevice = append(g.sentToDevice, userID)
	return nil
}
func (g *mockMessageGateway) BroadcastToRoom(ctx context.Context, roomID string, frameType int32, payload []byte) (int32, error) {
	g.broadcastRoom = append(g.broadcastRoom, roomID)
	return 1, nil
}
func (g *mockMessageGateway) BroadcastToGroup(ctx context.Context, groupID string, frameType int32, payload []byte) (int32, error) {
	g.broadcastGroup = append(g.broadcastGroup, groupID)
	return 1, nil
}

// mockUserStatusClient implements UserStatusClient for testing.
type mockUserStatusClient struct {
	online map[string][]OnlineDevice
}

func (c *mockUserStatusClient) GetUserOnline(ctx context.Context, userID string) ([]OnlineDevice, error) {
	return c.online[userID], nil
}

// mockSequenceClient implements pb.SequenceServiceClient for testing.
type mockSequenceClient struct {
	nextID int64
}

func (c *mockSequenceClient) NextBatchID(ctx context.Context, req *pb.NextBatchIDRequest, opts ...grpc.CallOption) (*pb.NextBatchIDResponse, error) {
	start := c.nextID + 1
	c.nextID += int64(req.Size)
	return &pb.NextBatchIDResponse{Start: start, End: c.nextID}, nil
}

func TestSendPrivateMessage(t *testing.T) {
	repo := &mockMessageRepo{}
	gw := &mockMessageGateway{}
	user := &mockUserStatusClient{online: map[string][]OnlineDevice{
		"receiver-1": {{DeviceID: "dev1", GatewayAddr: "localhost:9200"}},
	}}
	seq := &mockSequenceClient{}
	uc := NewSendUseCase(repo, gw, user, seq, nil)

	serverID, err := uc.SendPrivateMessage(context.Background(), "sender", "receiver-1", 0, "hello", "client-1", nil)
	if err != nil {
		t.Fatalf("SendPrivateMessage: %v", err)
	}
	if serverID == 0 {
		t.Fatal("expected non-zero serverID")
	}
}

func TestSendPrivateMessageOffline(t *testing.T) {
	repo := &mockMessageRepo{}
	gw := &mockMessageGateway{}
	user := &mockUserStatusClient{online: map[string][]OnlineDevice{}}
	seq := &mockSequenceClient{}
	uc := NewSendUseCase(repo, gw, user, seq, nil)

	serverID, err := uc.SendPrivateMessage(context.Background(), "sender", "offline-user", 0, "hi", "client-2", nil)
	if err != nil {
		t.Fatalf("SendPrivateMessage offline: %v", err)
	}
	if serverID == 0 {
		t.Fatal("expected serverID even for offline user")
	}
}

func TestSendChatroomMessage(t *testing.T) {
	repo := &mockMessageRepo{}
	gw := &mockMessageGateway{}
	user := &mockUserStatusClient{}
	seq := &mockSequenceClient{}
	uc := NewSendUseCase(repo, gw, user, seq, nil)

	serverID, err := uc.SendChatroomMessage(context.Background(), "sender", "room-1", 0, "hello room", "client-3", nil)
	if err != nil {
		t.Fatalf("SendChatroomMessage: %v", err)
	}
	if serverID == 0 {
		t.Fatal("expected non-zero serverID")
	}
}

func TestSendGroupMessage(t *testing.T) {
	repo := &mockMessageRepo{}
	gw := &mockMessageGateway{}
	user := &mockUserStatusClient{}
	seq := &mockSequenceClient{}
	uc := NewSendUseCase(repo, gw, user, seq, nil)

	serverID, err := uc.SendGroupMessage(context.Background(), "sender", "group-1", 0, "hello group", "client-4", nil)
	if err != nil {
		t.Fatalf("SendGroupMessage: %v", err)
	}
	if serverID == 0 {
		t.Fatal("expected non-zero serverID")
	}
}

func TestRecallMessage(t *testing.T) {
	repo := &mockMessageRepo{}
	gw := &mockMessageGateway{}
	user := &mockUserStatusClient{online: map[string][]OnlineDevice{
		"receiver-1": {{DeviceID: "dev1", GatewayAddr: "localhost:9200"}},
	}}
	seq := &mockSequenceClient{}
	uc := NewSendUseCase(repo, gw, user, seq, nil)

	// Insert a message first
	uc.SendPrivateMessage(context.Background(), "sender", "receiver-1", 0, "msg", "client-5", nil)

	err := uc.RecallMessage(context.Background(), 1, "sender")
	if err != nil {
		t.Fatalf("RecallMessage: %v", err)
	}
}

func TestAckMessage(t *testing.T) {
	repo := &mockMessageRepo{}
	gw := &mockMessageGateway{}
	user := &mockUserStatusClient{}
	seq := &mockSequenceClient{}
	uc := NewSendUseCase(repo, gw, user, seq, nil)

	err := uc.AckMessage(context.Background(), 1, "user-1")
	if err != nil {
		t.Fatalf("AckMessage: %v", err)
	}
}

func TestPullMessagesSince(t *testing.T) {
	repo := &mockMessageRepo{}
	gw := &mockMessageGateway{}
	user := &mockUserStatusClient{}
	seq := &mockSequenceClient{}
	uc := NewSendUseCase(repo, gw, user, seq, nil)

	// Insert some messages
	uc.SendPrivateMessage(context.Background(), "sender", "receiver-1", 0, "msg1", "c1", nil)
	uc.SendPrivateMessage(context.Background(), "sender", "receiver-1", 0, "msg2", "c2", nil)

	msgs, err := uc.PullMessagesSince(context.Background(), "receiver-1", 0, 50)
	if err != nil {
		t.Fatalf("PullMessagesSince: %v", err)
	}
	if len(msgs) != 2 {
		t.Fatalf("expected 2 messages, got %d", len(msgs))
	}
}

func TestAttachmentPropagation(t *testing.T) {
	repo := &mockMessageRepo{}
	gw := &mockMessageGateway{}
	user := &mockUserStatusClient{}
	seq := &mockSequenceClient{}
	uc := NewSendUseCase(repo, gw, user, seq, nil)

	att := &msgpb.Attachment{
		Type: &msgpb.Attachment_Image{
			Image: &msgpb.ImageAttachment{Url: "https://img/test.jpg", Width: 100, Height: 200},
		},
	}
	attBytes, _ := proto.Marshal(att)

	serverID, err := uc.SendPrivateMessage(context.Background(), "sender", "receiver-1", int32(msgpb.MessageType_MSG_TYPE_IMAGE), "photo", "client-att", attBytes)
	if err != nil {
		t.Fatalf("SendPrivateMessage with attachment: %v", err)
	}
	if serverID == 0 {
		t.Fatal("expected non-zero serverID for attachment message")
	}

	if len(repo.messages) != 1 {
		t.Fatal("expected 1 message in repo")
	}
	if len(repo.messages[0].Attachment) == 0 {
		t.Fatal("expected attachment data in stored message")
	}
}
