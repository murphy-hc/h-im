package data

import (
	"context"
	"fmt"

	"github.com/murphy-hc/h-im/services/sequence/internal/biz"
)

const defaultSegmentSize = 100

var _ biz.SequenceRepo = (*sequenceRepo)(nil)

type sequenceRepo struct {
	data *Data
}

// NewSequenceRepo creates a SequenceRepo implementation.
func NewSequenceRepo(data *Data) biz.SequenceRepo {
	return &sequenceRepo{data: data}
}

// AllocateSegment allocates a segment of IDs atomically.
func (r *sequenceRepo) AllocateSegment(ctx context.Context, key string, size int32) (start, end int64, step int32, err error) {
	if size <= 0 {
		size = defaultSegmentSize
	}

	tx, err := r.data.PG.Begin(ctx)
	if err != nil {
		return 0, 0, 0, fmt.Errorf("sequence: begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	// Auto-create sequence on first use — insert default row if key doesn't exist.
	_, err = tx.Exec(ctx,
		`INSERT INTO sequences (key, next_val, step, segment_size)
		 VALUES ($1, 1, 1, $2)
		 ON CONFLICT (key) DO NOTHING`,
		key, defaultSegmentSize,
	)
	if err != nil {
		return 0, 0, 0, fmt.Errorf("sequence: init key %s: %w", key, err)
	}

	// Atomically allocate segment.
	var startVal, endVal int64
	var stepVal int32
	err = tx.QueryRow(ctx,
		`UPDATE sequences
		 SET next_val = next_val + step * $2
		 WHERE key = $1
		 RETURNING next_val - step * $2, next_val - step, step`,
		key, size,
	).Scan(&startVal, &endVal, &stepVal)
	if err != nil {
		return 0, 0, 0, fmt.Errorf("sequence: allocate segment for %s: %w", key, err)
	}

	if err := tx.Commit(ctx); err != nil {
		return 0, 0, 0, fmt.Errorf("sequence: commit tx: %w", err)
	}

	return startVal, endVal, stepVal, nil
}
