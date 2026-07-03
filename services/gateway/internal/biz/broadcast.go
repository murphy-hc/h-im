package biz

import "context"

// BroadcastType constants matching proto ConversationType.
const (
	BroadcastTypeGroup = 1
	BroadcastTypeRoom  = 2
)

// BroadcastMsg is published to all gateway instances via Pub/Sub.
type BroadcastMsg struct {
	Type      int32  `json:"type"`
	TargetID  string `json:"target_id"`
	FrameType int32  `json:"frame_type"`
	Payload   []byte `json:"payload"`
	MsgID     string `json:"msg_id"`
}

// Broadcaster sends messages to all gateway instances (cross-instance fan-out).
type Broadcaster interface {
	Publish(ctx context.Context, msg *BroadcastMsg) error
}

// BroadcastListener subscribes to cross-gateway broadcast messages.
type BroadcastListener interface {
	// Subscribe starts listening. Messages are delivered to handler.
	// Call in a goroutine; blocks until Close is called.
	Subscribe(ctx context.Context, handler func(context.Context, *BroadcastMsg))
	// Close stops the subscription.
	Close() error
}
