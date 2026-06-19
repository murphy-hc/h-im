package biz

import "github.com/murphy-hc/h-im/pkg/snowflake"

// SequenceUseCase handles ID generation business logic.
type SequenceUseCase struct {
	gen *snowflake.Generator
}

// NewSequenceUseCase creates a SequenceUseCase.
func NewSequenceUseCase() *SequenceUseCase {
	gen := snowflake.New(1) // TODO: worker ID from config or service discovery
	return &SequenceUseCase{gen: gen}
}

// NextID returns the next unique ID.
func (uc *SequenceUseCase) NextID() int64 {
	return uc.gen.NextID()
}

// NextBatchID returns a batch of unique IDs.
func (uc *SequenceUseCase) NextBatchID(count int32) []int64 {
	ids := make([]int64, count)
	for i := int32(0); i < count; i++ {
		ids[i] = uc.gen.NextID()
	}
	return ids
}
