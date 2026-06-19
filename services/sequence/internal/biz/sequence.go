package biz

import (
	"context"
	"fmt"
)

// SequenceUseCase handles segment-based ID generation.
type SequenceUseCase struct {
	repo SequenceRepo
}

// NewSequenceUseCase creates a SequenceUseCase.
func NewSequenceUseCase(repo SequenceRepo) *SequenceUseCase {
	return &SequenceUseCase{repo: repo}
}

// AllocateSegment allocates a segment of IDs for the given key.
func (uc *SequenceUseCase) AllocateSegment(ctx context.Context, key string, size int32) (start, end int64, step int32, err error) {
	if key == "" {
		return 0, 0, 0, fmt.Errorf("key must not be empty")
	}
	return uc.repo.AllocateSegment(ctx, key, size)
}
