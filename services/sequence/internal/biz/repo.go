package biz

import "context"

// SequenceRepo defines the sequence data access interface.
type SequenceRepo interface {
	// AllocateSegment allocates a segment of IDs for the given key.
	// Returns the allocated range [start, end] and the step size.
	AllocateSegment(ctx context.Context, key string, size int32) (start, end int64, step int32, err error)
}
