package service

import (
	"context"

	"github.com/murphy-hc/h-im/services/gateway/internal/biz"
)

// PubSubService orchestrates broadcast lifecycle through biz interfaces.
// Server layer wraps this as transport.Server.
type PubSubService struct {
	broadcaster biz.Broadcaster
	listener    biz.BroadcastListener
}

// NewPubSubService creates a PubSubService.
func NewPubSubService(broadcaster biz.Broadcaster, listener biz.BroadcastListener) *PubSubService {
	return &PubSubService{broadcaster: broadcaster, listener: listener}
}

// Publish sends a broadcast message to all gateway instances.
func (s *PubSubService) Publish(ctx context.Context, msg *biz.BroadcastMsg) error {
	return s.broadcaster.Publish(ctx, msg)
}

// StartListening begins receiving broadcasts, dispatching to handler.
// Blocks until Close is called. Run in a goroutine.
func (s *PubSubService) StartListening(ctx context.Context, handler func(context.Context, *biz.BroadcastMsg)) {
	s.listener.Subscribe(ctx, handler)
}

// Close stops the broadcast listener.
func (s *PubSubService) Close() error {
	return s.listener.Close()
}
