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

// NextBatchID allocates a segment of IDs.
func (s *SequenceService) NextBatchID(ctx context.Context, req *pb.NextBatchIDRequest) (*pb.NextBatchIDResponse, error) {
	start, end, step, err := s.uc.AllocateSegment(ctx, req.GetKey(), req.GetSize())
	if err != nil {
		return nil, err
	}
	return &pb.NextBatchIDResponse{
		Start: start,
		End:   end,
		Step:  step,
	}, nil
}
