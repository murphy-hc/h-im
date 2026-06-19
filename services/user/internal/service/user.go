package service

import (
	pb "github.com/murphy-hc/h-im/gen/go/him/user/v1"
	"github.com/murphy-hc/h-im/services/user/internal/biz"
)

// UserService implements the UserService gRPC server.
type UserService struct {
	pb.UnimplementedUserServiceServer
	uc *biz.UserUseCase
}

// NewUserService creates a UserService.
func NewUserService(uc *biz.UserUseCase) *UserService {
	return &UserService{uc: uc}
}
