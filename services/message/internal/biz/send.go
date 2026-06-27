package biz

import (
	"context"
	"fmt"
	"sync/atomic"
	"time"

	gatewayv1 "github.com/murphy-hc/h-im/gen/go/him/gateway/v1"
	msgpb "github.com/murphy-hc/h-im/gen/go/him/message/v1"
	pb "github.com/murphy-hc/h-im/gen/go/him/sequence/v1"
	"github.com/murphy-hc/h-im/pkg/gp"
	"golang.org/x/sync/singleflight"
	"google.golang.org/protobuf/proto"
)

const (
	maxRetries    = 3
	baseDelay     = 500 * time.Millisecond
	pushTimeout   = 10 * time.Second
	goroutinePool = 256
	defaultBatch  = 100
)

var pushSem = make(chan struct{}, goroutinePool)

// idAllocator pre-allocates message IDs in batches from the sequence service.
type idAllocator struct {
	seq       pb.SequenceServiceClient
	sf        singleflight.Group
	current   atomic.Int64
	end       atomic.Int64
	batchSize int32
}

func newIDAllocator(seq pb.SequenceServiceClient) *idAllocator {
	return &idAllocator{seq: seq, batchSize: defaultBatch}
}

func (a *idAllocator) NextID(ctx context.Context) (int64, error) {
	for {
		cur := a.current.Load()
		end := a.end.Load()

		if cur > end || end-cur <= int64(float64(a.batchSize)*0.1) {
			_, err, _ := a.sf.Do("id", func() (any, error) {
				cur2 := a.current.Load()
				end2 := a.end.Load()
				if cur2 <= end2 && end2-cur2 > int64(float64(a.batchSize)*0.1) {
					return nil, nil
				}
				resp, err := a.seq.NextBatchID(ctx, &pb.NextBatchIDRequest{
					Key: "private_message_id", Size: a.batchSize,
				})
				if err != nil {
					return nil, err
				}
				a.current.Store(resp.Start)
				a.end.Store(resp.End)
				return nil, nil
			})
			if err != nil {
				return 0, err
			}
			continue
		}

		next := cur + 1
		if a.current.CompareAndSwap(cur, next) {
			return cur, nil
		}
	}
}

type SendUseCase struct {
	repo MessageRepo
	gw   MessageGateway
	user UserStatusClient
	id   *idAllocator
}

func NewSendUseCase(repo MessageRepo, gw MessageGateway, user UserStatusClient, seq pb.SequenceServiceClient) *SendUseCase {
	return &SendUseCase{repo: repo, gw: gw, user: user, id: newIDAllocator(seq)}
}

func (uc *SendUseCase) SendPrivateMessage(ctx context.Context, senderID, receiverID string, msgType int32, text, clientID string) (int64, error) {
	serverID, err := uc.id.NextID(ctx)
	if err != nil {
		return 0, fmt.Errorf("get sequence: %w", err)
	}
	now := time.Now().UnixMilli()
	m := &Message{
		ServerID:   serverID,
		ClientID:   clientID,
		SenderID:   senderID,
		ReceiverID: receiverID,
		MsgType:    msgType,
		Text:       text,
		ServerTime: now,
		CreateTime: now,
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

// RecallMessage marks a message as recalled and pushes a recall notification.
func (uc *SendUseCase) RecallMessage(ctx context.Context, serverID int64, senderID string) error {
	ok, err := uc.repo.MarkRecalled(ctx, serverID, senderID, time.Now().UnixMilli())
	if err != nil {
		return err
	}
	if !ok {
		return fmt.Errorf("recall denied")
	}

	receiverID, err := uc.repo.GetReceiverID(ctx, serverID)
	if err != nil {
		return err
	}

	push, _ := proto.Marshal(&msgpb.RecallPush{
		MessageServerId: serverID,
		RecallTime:      time.Now().UnixMilli(),
	})
	gp.SafeGo(ctx, func(pushCtx context.Context) {
		uc.pushToDevices(pushCtx, receiverID, int32(gatewayv1.FrameType_FRAME_TYPE_PRIVATE_RECALL), push, nil, 0)
	})
	return nil
}

func (uc *SendUseCase) AckMessage(ctx context.Context, serverID int64, userID string) error {
	return uc.repo.MarkRead(ctx, serverID)
}

// PullMessagesSince returns messages for a user since a given message ID.
func (uc *SendUseCase) PullMessagesSince(ctx context.Context, userID string, sinceID int64, limit int32) ([]Message, error) {
	return uc.repo.PullSince(ctx, userID, sinceID, limit)
}

// pushToDevices delivers a payload to all online devices of a user.
func (uc *SendUseCase) pushToDevices(ctx context.Context, userID string, frameType int32, payload []byte, markDelivered func(ctx context.Context, serverID int64) error, msgServerID int64) {
	ctx, cancel := context.WithTimeout(ctx, pushTimeout)
	defer cancel()
	devices, err := uc.user.GetUserOnline(ctx, userID)
	if err != nil || len(devices) == 0 {
		return
	}
	pushSem <- struct{}{}
	defer func() { <-pushSem }()

	delivered := false
	for _, device := range devices {
		for i := 0; i < maxRetries; i++ {
			err := uc.gw.SendToDevice(ctx, device.GatewayAddr, userID, frameType, payload)
			if err == nil {
				delivered = true
				break
			}
			select {
			case <-ctx.Done():
				return
			case <-time.After(baseDelay * time.Duration(1<<i)):
			}
		}
	}
	if delivered && markDelivered != nil {
		markDelivered(ctx, msgServerID)
	}
}

func (uc *SendUseCase) pushToReceiver(m *Message) {
	payload, _ := proto.Marshal(&msgpb.Message{
		MessageClientId: m.ClientID,
		MessageServerId: m.ServerID,
		SenderId:        m.SenderID,
		ReceiverId:      m.ReceiverID,
		Text:            m.Text,
		ServerTime:      m.ServerTime,
		MsgType:         msgpb.MessageType(m.MsgType),
		ConvType:        msgpb.ConversationType_CONVERSATION_PRIVATE,
	})
	uc.pushToDevices(context.Background(), m.ReceiverID, int32(gatewayv1.FrameType_FRAME_TYPE_PRIVATE_CHAT), payload, uc.repo.MarkDelivered, m.ServerID)
}
