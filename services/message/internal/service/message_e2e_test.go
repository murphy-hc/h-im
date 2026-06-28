package service

import (
	"context"
	"sync"
	"testing"

	pb "github.com/murphy-hc/h-im/gen/go/him/message/v1"
	seqpb "github.com/murphy-hc/h-im/gen/go/him/sequence/v1"
	"github.com/murphy-hc/h-im/services/message/internal/biz"
	"google.golang.org/grpc"
)

// ---- mocks (same pattern as biz tests) ----

type e2eMsgRepo struct {
	mu       sync.Mutex
	messages []*biz.Message
}

func (r *e2eMsgRepo) Insert(_ context.Context, m *biz.Message) error {
	r.mu.Lock(); defer r.mu.Unlock()
	r.messages = append(r.messages, m)
	return nil
}
func (r *e2eMsgRepo) InsertChatroom(_ context.Context, _ int64, _, _, _ string, _ int32, _, _ string, _ int64) error {
	return nil
}
func (r *e2eMsgRepo) InsertGroup(_ context.Context, _ int64, _, _, _ string, _ int32, _, _ string, _ int64) error {
	return nil
}
func (r *e2eMsgRepo) GetReceiverID(_ context.Context, serverID int64) (string, error) { return "receiver", nil }
func (r *e2eMsgRepo) MarkDelivered(_ context.Context, _ int64) error                  { return nil }
func (r *e2eMsgRepo) MarkRead(_ context.Context, _ int64) error                       { return nil }
func (r *e2eMsgRepo) MarkRecalled(_ context.Context, _ int64, _ string, _ int64) (bool, error) {
	return true, nil
}
func (r *e2eMsgRepo) PullSince(_ context.Context, _ string, _ int64, _ int32) ([]biz.Message, error) {
	r.mu.Lock(); defer r.mu.Unlock()
	out := make([]biz.Message, len(r.messages))
	for i, m := range r.messages {
		out[i] = *m
	}
	return out, nil
}

type e2eMsgGW struct{}

func (g *e2eMsgGW) SendToDevice(_ context.Context, _, _ string, _ int32, _ []byte) error { return nil }
func (g *e2eMsgGW) BroadcastToRoom(_ context.Context, _ string, _ int32, _ []byte) (int32, error) {
	return 0, nil
}
func (g *e2eMsgGW) BroadcastToGroup(_ context.Context, _ string, _ int32, _ []byte) (int32, error) {
	return 0, nil
}

type e2eUserCli struct{}

func (c *e2eUserCli) GetUserOnline(_ context.Context, _ string) ([]biz.OnlineDevice, error) {
	return nil, nil
}

type e2eSeqCli struct {
	next int64
}

func (c *e2eSeqCli) NextBatchID(_ context.Context, req *seqpb.NextBatchIDRequest, _ ...grpc.CallOption) (*seqpb.NextBatchIDResponse, error) {
	start := c.next + 1
	c.next += int64(req.Size)
	return &seqpb.NextBatchIDResponse{Start: start, End: c.next}, nil
}

func newE2EMessageService() *MessageService {
	repo := &e2eMsgRepo{}
	gw := &e2eMsgGW{}
	user := &e2eUserCli{}
	seq := &e2eSeqCli{}
	uc := biz.NewSendUseCase(repo, gw, user, seq, nil)
	return NewMessageService(uc)
}

func TestE2E_SendAckPullRecall(t *testing.T) {
	svc := newE2EMessageService()
	ctx := context.Background()

	// 1. Send message
	sendResp, err := svc.SendMessage(ctx, &pb.SendMessageReq{
		SenderId:        "alice",
		ReceiverId:      "bob",
		ConvType:        pb.ConversationType_CONVERSATION_PRIVATE,
		MsgType:         pb.MessageType_MSG_TYPE_TEXT,
		Text:            "Hello Bob",
		MessageClientId: "client-001",
	})
	if err != nil {
		t.Fatalf("SendMessage: %v", err)
	}
	serverID := sendResp.MessageServerId
	if serverID == 0 {
		t.Fatal("expected non-zero server ID")
	}

	// 2. Ack message
	_, err = svc.AckMessage(ctx, &pb.AckMessageReq{
		MessageServerId: serverID,
		UserId:          "bob",
	})
	if err != nil {
		t.Fatalf("AckMessage: %v", err)
	}

	// 3. Pull messages
	pullResp, err := svc.PullMessages(ctx, &pb.PullMessagesReq{
		UserId:         "bob",
		SinceMessageId: 0,
		Limit:          50,
	})
	if err != nil {
		t.Fatalf("PullMessages: %v", err)
	}
	if len(pullResp.Messages) == 0 {
		t.Fatal("expected at least 1 message")
	}

	// 4. Recall
	_, err = svc.RecallMessage(ctx, &pb.RecallMessageReq{
		MessageServerId: serverID,
		SenderId:        "alice",
	})
	if err != nil {
		t.Fatalf("RecallMessage: %v", err)
	}
}

func TestE2E_SendToChatroom(t *testing.T) {
	svc := newE2EMessageService()
	ctx := context.Background()

	resp, err := svc.SendMessage(ctx, &pb.SendMessageReq{
		SenderId:        "alice",
		ReceiverId:      "room-1",
		ConvType:        pb.ConversationType_CONVERSATION_ROOM,
		MsgType:         pb.MessageType_MSG_TYPE_TEXT,
		Text:            "Hello Room",
		MessageClientId: "client-room-001",
	})
	if err != nil {
		t.Fatalf("SendMessage to room: %v", err)
	}
	if resp.MessageServerId == 0 {
		t.Fatal("expected non-zero server ID")
	}
}

func TestE2E_SendToGroup(t *testing.T) {
	svc := newE2EMessageService()
	ctx := context.Background()

	resp, err := svc.SendMessage(ctx, &pb.SendMessageReq{
		SenderId:        "alice",
		ReceiverId:      "group-1",
		ConvType:        pb.ConversationType_CONVERSATION_GROUP,
		MsgType:         pb.MessageType_MSG_TYPE_TEXT,
		Text:            "Hello Group",
		MessageClientId: "client-group-001",
	})
	if err != nil {
		t.Fatalf("SendMessage to group: %v", err)
	}
	if resp.MessageServerId == 0 {
		t.Fatal("expected non-zero server ID")
	}
}
