package server

import (
	"context"

	"github.com/murphy-hc/h-im/services/gateway/internal/biz"
	"github.com/murphy-hc/h-im/services/gateway/internal/service"
)

// PubSubServer wraps service.PubSubService as a kratos transport.Server.
type PubSubServer struct {
	svc     *service.PubSubService
	handler func(context.Context, *biz.BroadcastMsg)
}

// NewPubSubServer creates a transport server for broadcast lifecycle management.
func NewPubSubServer(svc *service.PubSubService, grpcSvc *service.GatewayGrpcService) *PubSubServer {
	return &PubSubServer{
		svc:     svc,
		handler: grpcSvc.HandleBroadcast,
	}
}

// Start begins listening for cross-gateway broadcasts.
func (s *PubSubServer) Start(ctx context.Context) error {
	go s.svc.StartListening(ctx, s.handler)
	return nil
}

// Stop shuts down the broadcast listener.
func (s *PubSubServer) Stop(ctx context.Context) error {
	return s.svc.Close()
}
