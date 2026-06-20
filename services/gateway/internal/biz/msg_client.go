package biz

import (
	"context"

	msgpb "github.com/murphy-hc/h-im/gen/go/him/message/v1"
)

// MessageClient is the interface for calling the Message service.
type MessageClient interface {
	SendMessage(ctx context.Context, req *msgpb.SendMessageReq) (*msgpb.SendMessageResp, error)
	AckMessage(ctx context.Context, serverID int64, userID string) error
}
