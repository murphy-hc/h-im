package data

import (
	"context"
	"fmt"

	"github.com/murphy-hc/h-im/services/sequence/internal/biz"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const defaultSegmentSize = 100

var _ biz.SequenceRepo = (*sequenceRepo)(nil)

// SequenceModel is the GORM model for the sequences table.
type SequenceModel struct {
	Key         string `gorm:"primaryKey;size:128"`
	NextVal     int64  `gorm:"column:next_val;default:1"`
	Step        int32  `gorm:"column:step;default:1"`
	SegmentSize int32  `gorm:"column:segment_size;default:100"`
}

// TableName overrides the default table name.
func (SequenceModel) TableName() string {
	return "sequences"
}

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

	err = r.data.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Ensure the key exists — insert default row on first use.
		if err := tx.Clauses(clause.OnConflict{DoNothing: true}).Create(&SequenceModel{
			Key:         key,
			NextVal:     1,
			Step:        1,
			SegmentSize: defaultSegmentSize,
		}).Error; err != nil {
			return fmt.Errorf("sequence: init key %s: %w", key, err)
		}

		// Lock the row for update and read current values.
		var m SequenceModel
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("`key` = ?", key).
			First(&m).Error; err != nil {
			return fmt.Errorf("sequence: read key %s: %w", key, err)
		}

		// Compute the allocated range.
		incr := int64(m.Step) * int64(size)
		start = m.NextVal
		end = m.NextVal + incr - int64(m.Step)
		step = m.Step

		// Atomically advance.
		if err := tx.Model(&m).Update("next_val", m.NextVal+incr).Error; err != nil {
			return fmt.Errorf("sequence: advance key %s: %w", key, err)
		}

		return nil
	})
	if err != nil {
		return 0, 0, 0, err
	}

	return start, end, step, nil
}
