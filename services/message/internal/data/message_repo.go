package data

import (
	"context"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// MessageRepo provides message persistence.
type MessageRepo struct {
	db *gorm.DB
}

// NewMessageRepo creates a MessageRepo.
func NewMessageRepo(data *Data) *MessageRepo {
	return &MessageRepo{db: data.DB}
}

// Insert inserts a message, silently skipping duplicates (idempotent via client_id UNIQUE).
func (r *MessageRepo) Insert(ctx context.Context, m *MessageModel) error {
	return r.db.WithContext(ctx).Clauses(clause.OnConflict{DoNothing: true}).Create(m).Error
}

// MarkRemoteRead marks a message as remotely read.
func (r *MessageRepo) MarkRemoteRead(ctx context.Context, serverID int64) error {
	return r.db.WithContext(ctx).Model(&MessageModel{}).
		Where("message_server_id = ?", serverID).Update("is_remote_read", true).Error
}
