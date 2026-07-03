package data

import (
	"context"
	"encoding/json"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/murphy-hc/h-im/services/gateway/internal/biz"
	"github.com/redis/go-redis/v9"
)

const broadcastChannel = "gw:broadcast"

// PubSub provides Redis Pub/Sub operations for cross-gateway broadcast.
// Only raw Redis ops: Subscribe, Publish, Close. Lifecycle managed by server layer.
// Implements biz.Broadcaster.
type PubSub struct {
	rdb    *redis.Client
	pubsub *redis.PubSub
	stopCh chan struct{}
}

// NewPubSub creates a PubSub. Pass handler to Subscribe to start listening.
func NewPubSub(rdb *redis.Client) *PubSub {
	return &PubSub{rdb: rdb, stopCh: make(chan struct{})}
}

// Subscribe starts listening on the broadcast channel. Messages are delivered to handler.
// Blocks until stopCh is closed. Call in a goroutine.
func (ps *PubSub) Subscribe(ctx context.Context, handler func(context.Context, *biz.BroadcastMsg)) {
	ps.pubsub = ps.rdb.Subscribe(ctx, broadcastChannel)
	ch := ps.pubsub.Channel()
	for {
		select {
		case <-ps.stopCh:
			return
		case msg, ok := <-ch:
			if !ok {
				return
			}
			var bm biz.BroadcastMsg
			if err := json.Unmarshal([]byte(msg.Payload), &bm); err != nil {
				log.Warnf("pubsub: unmarshal: %v", err)
				continue
			}
			if bm.MsgID == "" {
				continue
			}
			handler(ctx, &bm)
		}
	}
}

// Publish sends a broadcast message to all gateway instances.
func (ps *PubSub) Publish(ctx context.Context, msg *biz.BroadcastMsg) error {
	data, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	return ps.rdb.Publish(ctx, broadcastChannel, string(data)).Err()
}

// Close stops the subscriber and cleans up.
func (ps *PubSub) Close() error {
	close(ps.stopCh)
	if ps.pubsub != nil {
		return ps.pubsub.Close()
	}
	return nil
}
