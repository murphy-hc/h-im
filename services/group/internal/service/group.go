package service

import (
	pb "github.com/murphy-hc/h-im/gen/go/him/group/v1"
	"github.com/murphy-hc/h-im/services/group/internal/biz"
)

// GroupService implements the GroupService gRPC server.
type GroupService struct {
	pb.UnimplementedGroupServiceServer
	uc *biz.GroupUseCase
}

// NewGroupService creates a GroupService.
func NewGroupService(uc *biz.GroupUseCase) *GroupService {
	return &GroupService{uc: uc}
}
