package biz

import (
	"context"
	"fmt"
	"time"

	gatewayv1 "github.com/murphy-hc/h-im/gen/go/him/gateway/v1"
	msgpb "github.com/murphy-hc/h-im/gen/go/him/message/v1"
	pb "github.com/murphy-hc/h-im/gen/go/him/sequence/v1"
	"github.com/murphy-hc/h-im/pkg/gp"
	"github.com/murphy-hc/h-im/services/message/internal/data"
	"google.golang.org/protobuf/proto"
)

const (
	maxRetries    = 3
	baseDelay     = 500 * time.Millisecond
	pushTimeout   = 10 * time.Second
	goroutinePool = 256
)

var pushSem = make(chan struct{}, goroutinePool)

type SendUseCase struct {
	repo *data.MessageRepo
	seq  pb.SequenceServiceClient
	gw   *data.GatewayClient
	user *data.UserClient
}

func NewSendUseCase(repo *data.MessageRepo, seq pb.SequenceServiceClient, gw *data.GatewayClient, user *data.UserClient) *SendUseCase {
	return &SendUseCase{repo: repo, seq: seq, gw: gw, user: user}
}

func (uc *SendUseCase) SendPrivateMessage(ctx context.Context, senderID, receiverID string, msgType int32, text, clientID string) (int64, error) {
	resp, err := uc.seq.NextBatchID(ctx, &pb.NextBatchIDRequest{Key: "private_message_id", Size: 1})
	if err != nil {
		return 0, fmt.Errorf("get sequence: %w", err)
	}
	serverID := resp.Start
	now := time.Now().UnixMilli()
	m := &data.MessageModel{
		MessageServerID: serverID,
		MessageClientID: clientID,
		SenderID:        senderID,
		ReceiverID:      receiverID,
		ConvType:        0,
		MsgType:         msgType,
		Text:            text,
		ServerTime:      now,
		CreateTime:      now,
	}
	if err := uc.repo.Insert(ctx, m); err != nil {
		return 0, fmt.Errorf("insert message: %w", err)
	}
	gp.SafeGo(ctx, func(_ context.Context) {
		pushSem <- struct{}{}
		defer func() { <-pushSem }()
		uc.pushToReceiver(m)
	})
	return serverID, nil
}

func (uc *SendUseCase) AckMessage(ctx context.Context, serverID int64, userID string) error {
	return uc.repo.MarkRemoteRead(ctx, serverID)
}

// PullMessagesSince returns messages for a user since a given message ID.
func (uc *SendUseCase) PullMessagesSince(ctx context.Context, userID string, sinceID int64, limit int32) ([]data.MessageModel, error) {
	return uc.repo.PullMessagesSince(ctx, userID, sinceID, limit)
}

func (uc *SendUseCase) pushToReceiver(m *data.MessageModel) {
	payload, _ := proto.Marshal(&msgpb.Message{
		MessageClientId: m.MessageClientID,
		MessageServerId: m.MessageServerID,
		SenderId:        m.SenderID,
		ReceiverId:      m.ReceiverID,
		Text:            m.Text,
		ServerTime:      m.ServerTime,
		MsgType:         msgpb.MessageType(m.MsgType),
		ConvType:        msgpb.ConversationType_CONVERSATION_PRIVATE,
	})

	ctx, cancel := context.WithTimeout(context.Background(), pushTimeout)
	defer cancel()

	// Check online status via user service
	onlineResp, err := uc.user.GetUserOnline(ctx, m.ReceiverID)
	if err != nil || len(onlineResp.GetDevices()) == 0 {
		return
	}

	// Deliver to each online device via direct gateway connection
	for _, device := range onlineResp.GetDevices() {
		for i := 0; i < maxRetries; i++ {
			err := uc.gw.SendToUserDirect(ctx, device.GatewayAddr, m.ReceiverID,
				int32(gatewayv1.FrameType_FRAME_TYPE_PRIVATE_CHAT), payload)
			if err == nil {
				break
			}
			select {
			case <-ctx.Done():
				return
			case <-time.After(baseDelay * time.Duration(1<<i)):
			}
		}
	}
}
