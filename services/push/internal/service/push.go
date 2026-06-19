package service

import (
	pb "github.com/murphy-hc/h-im/gen/go/him/push/v1"
	"github.com/murphy-hc/h-im/services/push/internal/biz"
)

// PushService implements the PushService gRPC server.
type PushService struct {
	pb.UnimplementedPushServiceServer
	uc *biz.PushUseCase
}

// NewPushService creates a PushService.
func NewPushService(uc *biz.PushUseCase) *PushService {
	return &PushService{uc: uc}
}
