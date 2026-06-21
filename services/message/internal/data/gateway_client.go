package data

import (
	"context"
	"sync"

	kgrpc "github.com/go-kratos/kratos/v2/transport/grpc"
	gwpb "github.com/murphy-hc/h-im/gen/go/him/gateway/v1"
	ggrpc "google.golang.org/grpc"
)

// GatewayClient proxies calls to the Gateway service.
type GatewayClient struct {
	client gwpb.GatewayServiceClient

	mu   sync.RWMutex
	pool map[string]gwpb.GatewayServiceClient // gatewayAddr -> client
}

// NewGatewayClient creates a Kratos gRPC client for the Gateway service.
func NewGatewayClient() (*GatewayClient, func(), error) {
	var conns []*ggrpc.ClientConn
	conn, err := kgrpc.DialInsecure(context.Background(),
		kgrpc.WithEndpoint("discovery:///gateway.default.svc.cluster.local:9200"),
	)
	if err != nil {
		return nil, nil, err
	}
	conns = append(conns, conn)
	return &GatewayClient{
		client: gwpb.NewGatewayServiceClient(conn),
		pool:   make(map[string]gwpb.GatewayServiceClient),
	}, func() {
		for _, c := range conns {
			c.Close()
		}
	}, nil
}

// SendToUserDirect sends a message to a specific gateway instance.
func (c *GatewayClient) SendToUserDirect(ctx context.Context, gatewayAddr string, userID string, frameType int32, payload []byte) error {
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

// getOrDial returns a cached gRPC client for the given gateway address, or dials a new one.
func (c *GatewayClient) getOrDial(addr string) (gwpb.GatewayServiceClient, error) {
	c.mu.RLock()
	cl, ok := c.pool[addr]
	c.mu.RUnlock()
	if ok {
		return cl, nil
	}

	c.mu.Lock()
	defer c.mu.Unlock()
	if cl, ok = c.pool[addr]; ok {
		return cl, nil
	}

	conn, err := kgrpc.DialInsecure(context.Background(),
		kgrpc.WithEndpoint(addr),
	)
	if err != nil {
		return nil, err
	}
	cl = gwpb.NewGatewayServiceClient(conn)
	c.pool[addr] = cl
	return cl, nil
}
