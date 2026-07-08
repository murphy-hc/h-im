package data

import (
	"context"
	"sync"
	"time"

	kgrpc "github.com/go-kratos/kratos/v2/transport/grpc"
	gwpb "github.com/murphy-hc/h-im/gen/go/him/gateway/v1"
	"github.com/murphy-hc/h-im/services/message/internal/biz"
	ggrpc "google.golang.org/grpc"
)

var _ biz.MessageGateway = (*GatewayClient)(nil)

const poolCleanupInterval = 5 * time.Minute

type poolConn struct {
	client   gwpb.GatewayServiceClient
	conn     *ggrpc.ClientConn
	lastUsed time.Time
}

// GatewayClient proxies calls to the Gateway service.
type GatewayClient struct {
	client gwpb.GatewayServiceClient

	mu      sync.RWMutex
	pool    map[string]*poolConn // gatewayAddr -> conn
	conns   []*ggrpc.ClientConn  // all connections for cleanup on shutdown
	closeCh chan struct{}
}

// NewGatewayClient creates a Kratos gRPC client for the Gateway service.
func NewGatewayClient() (*GatewayClient, func(), error) {
	conn, err := kgrpc.DialInsecure(context.Background(),
		kgrpc.WithEndpoint("discovery:///gateway.default.svc.cluster.local:9200"),
	)
	if err != nil {
		return nil, nil, err
	}
	gc := &GatewayClient{
		client:  gwpb.NewGatewayServiceClient(conn),
		pool:    make(map[string]*poolConn),
		conns:   []*ggrpc.ClientConn{conn},
		closeCh: make(chan struct{}),
	}
	go gc.cleanupLoop()
	return gc, func() {
		close(gc.closeCh)
		gc.mu.Lock()
		for _, pc := range gc.pool {
			pc.conn.Close()
		}
		gc.mu.Unlock()
		for _, c := range gc.conns {
			c.Close()
		}
	}, nil
}

func (c *GatewayClient) cleanupLoop() {
	ticker := time.NewTicker(poolCleanupInterval)
	defer ticker.Stop()
	for {
		select {
		case <-c.closeCh:
			return
		case <-ticker.C:
			c.evictIdle()
		}
	}
}

func (c *GatewayClient) evictIdle() {
	c.mu.Lock()
	defer c.mu.Unlock()
	cutoff := time.Now().Add(-poolCleanupInterval)
	for addr, pc := range c.pool {
		if pc.lastUsed.Before(cutoff) {
			pc.conn.Close()
			delete(c.pool, addr)
		}
	}
}

// SendToDevice sends a message to a specific gateway instance.
func (c *GatewayClient) SendToDevice(ctx context.Context, gatewayAddr, userID string, frameType int32, payload []byte) error {
	cl, err := c.getOrDial(gatewayAddr)
	if err != nil {
		return err
	}
	_, err = cl.SendToUser(ctx, &gwpb.SendToUserRequest{
		UserId:    userID,
		FrameType: frameType,
		Payload:   payload,
	})
	return err
}

// BroadcastToRoom sends a message to all members of a chatroom via the gateway.
func (c *GatewayClient) BroadcastToRoom(ctx context.Context, roomID string, frameType int32, payload []byte) (int32, error) {
	resp, err := c.client.BroadcastToChatroom(ctx, &gwpb.BroadcastToChatroomRequest{
		RoomId:    roomID,
		FrameType: frameType,
		Payload:   payload,
	})
	if err != nil {
		return 0, err
	}
	return resp.DeliveredCount, nil
}

// BroadcastToGroup sends a message to all members of a group via the gateway.
func (c *GatewayClient) BroadcastToGroup(ctx context.Context, groupID string, frameType int32, payload []byte) (int32, error) {
	resp, err := c.client.BroadcastToGroup(ctx, &gwpb.BroadcastToGroupRequest{
		GroupId:   groupID,
		FrameType: frameType,
		Payload:   payload,
	})
	if err != nil {
		return 0, err
	}
	return resp.DeliveredCount, nil
}

// getOrDial returns a cached gRPC client for the given gateway address, or dials a new one.
func (c *GatewayClient) getOrDial(addr string) (gwpb.GatewayServiceClient, error) {
	c.mu.RLock()
	pc, ok := c.pool[addr]
	c.mu.RUnlock()
	if ok {
		pc.lastUsed = time.Now()
		return pc.client, nil
	}

	c.mu.Lock()
	defer c.mu.Unlock()
	if pc, ok = c.pool[addr]; ok {
		pc.lastUsed = time.Now()
		return pc.client, nil
	}

	conn, err := kgrpc.DialInsecure(context.Background(),
		kgrpc.WithEndpoint(addr),
	)
	if err != nil {
		return nil, err
	}
	pc = &poolConn{
		client:   gwpb.NewGatewayServiceClient(conn),
		conn:     conn,
		lastUsed: time.Now(),
	}
	c.pool[addr] = pc
	c.conns = append(c.conns, conn)
	return pc.client, nil
}
