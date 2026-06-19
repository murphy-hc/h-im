package service

import (
	pb "github.com/murphy-hc/h-im/gen/go/him/media/v1"
	"github.com/murphy-hc/h-im/services/media/internal/biz"
)

// MediaService implements the MediaService gRPC server.
type MediaService struct {
	pb.UnimplementedMediaServiceServer
	uc *biz.MediaUseCase
}

// NewMediaService creates a MediaService.
func NewMediaService(uc *biz.MediaUseCase) *MediaService {
	return &MediaService{uc: uc}
}
