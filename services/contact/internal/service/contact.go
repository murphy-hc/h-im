package service

import (
	pb "github.com/murphy-hc/h-im/gen/go/him/contact/v1"
	"github.com/murphy-hc/h-im/services/contact/internal/biz"
)

// ContactService implements the ContactService gRPC server.
type ContactService struct {
	pb.UnimplementedContactServiceServer
	uc *biz.ContactUseCase
}

// NewContactService creates a ContactService.
func NewContactService(uc *biz.ContactUseCase) *ContactService {
	return &ContactService{uc: uc}
}
