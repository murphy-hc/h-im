package service

import (
	pb "github.com/murphy-hc/h-im/gen/go/him/auth/v1"
	"github.com/murphy-hc/h-im/services/auth/internal/biz"
)

// AuthService implements the AuthService gRPC server.
type AuthService struct {
	pb.UnimplementedAuthServiceServer
	uc *biz.AuthUseCase
}

// NewAuthService creates a AuthService.
func NewAuthService(uc *biz.AuthUseCase) *AuthService {
	return &AuthService{uc: uc}
}
