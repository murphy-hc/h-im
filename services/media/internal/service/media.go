package service

import (
	"context"

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

func (s *MediaService) UploadMedia(ctx context.Context, req *pb.UploadMediaRequest) (*pb.UploadMediaResponse, error) {
	return &pb.UploadMediaResponse{}, nil
}

func (s *MediaService) DownloadMedia(ctx context.Context, req *pb.DownloadMediaRequest) (*pb.DownloadMediaResponse, error) {
	m, err := s.uc.GetMedia(ctx, req.MediaId)
	if err != nil {
		return nil, err
	}
	return &pb.DownloadMediaResponse{
		Data:     nil, // caller uses the URL from GetMediaInfo to download directly
		MimeType: m.MimeType,
	}, nil
}

func (s *MediaService) GetMediaInfo(ctx context.Context, req *pb.GetMediaInfoRequest) (*pb.GetMediaInfoResponse, error) {
	m, err := s.uc.GetMedia(ctx, req.MediaId)
	if err != nil {
		return nil, err
	}
	return &pb.GetMediaInfoResponse{
		Media: &pb.MediaInfo{
			MediaId:   m.MediaID,
			Url:       m.URL,
			ThumbUrl:  m.ThumbURL,
			Type:      pb.MediaType(m.MediaType),
			MimeType:  m.MimeType,
			Size:      m.Size,
			CreatedAt: 0,
		},
	}, nil
}
