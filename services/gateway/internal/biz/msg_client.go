package biz

import (
	"context"

	msgpb "github.com/murphy-hc/h-im/gen/go/him/message/v1"
)

// MessageClient is the interface for calling the Message service.
type MessageClient interface {
	SendMessage(ctx context.Context, req *msgpb.SendMessageReq) (*msgpb.SendMessageResp, error)
	AckMessage(ctx context.Context, serverID int64, userID string) error
	RecallMessage(ctx context.Context, req *msgpb.RecallMessageReq) error
	PullMessages(ctx context.Context, userID string, sinceID int64, limit int32) ([]PullMessage, error)
}

// PullMessage is a simplified message for pull results.
type PullMessage struct {
	ClientID   string
	ServerID   int64
	SenderID   string
	ReceiverID string
	ConvType   int32
	MsgType    int32
	Text       string
	ServerTime int64
}
