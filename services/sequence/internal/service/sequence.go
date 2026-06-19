package service

import (
	"context"

	pb "github.com/murphy-hc/h-im/gen/go/him/sequence/v1"
	"github.com/murphy-hc/h-im/services/sequence/internal/biz"
)

// SequenceService implements the SequenceService gRPC server.
type SequenceService struct {
	pb.UnimplementedSequenceServiceServer
	uc *biz.SequenceUseCase
}

// NewSequenceService creates a SequenceService.
func NewSequenceService(uc *biz.SequenceUseCase) *SequenceService {
	return &SequenceService{uc: uc}
}

// NextID returns a single snowflake ID.
func (s *SequenceService) NextID(ctx context.Context, req *pb.NextIDRequest) (*pb.NextIDResponse, error) {
	return &pb.NextIDResponse{Id: s.uc.NextID()}, nil
}

// NextBatchID returns a batch of snowflake IDs.
func (s *SequenceService) NextBatchID(ctx context.Context, req *pb.NextBatchIDRequest) (*pb.NextBatchIDResponse, error) {
	count := req.GetCount()
	ids := s.uc.NextBatchID(count)
	return &pb.NextBatchIDResponse{Ids: ids}, nil
}
