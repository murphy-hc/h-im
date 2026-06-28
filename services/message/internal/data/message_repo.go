package data

import (
	"context"

	"github.com/murphy-hc/h-im/services/message/internal/biz"
	msgpb "github.com/murphy-hc/h-im/gen/go/him/message/v1"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var _ biz.MessageRepo = (*MessageRepo)(nil)

// MessageRepo provides message persistence.
type MessageRepo struct {
	db *gorm.DB
}

// NewMessageRepo creates a MessageRepo.
func NewMessageRepo(data *Data) *MessageRepo {
	return &MessageRepo{db: data.DB}
}

// InsertChatroom inserts a chatroom message.
func (r *MessageRepo) InsertChatroom(ctx context.Context, serverID int64, clientID, roomID, senderID string, msgType int32, text, attachment string, serverTime int64) error {
	return r.db.WithContext(ctx).Create(&ChatroomMessageModel{
		MessageServerID: serverID, MessageClientID: clientID, RoomID: roomID,
		SenderID: senderID, MsgType: msgType, Text: text, Attachment: attachment,
		ServerTime: serverTime, Status: 1,
	}).Error
}

// InsertGroup inserts a group message.
func (r *MessageRepo) InsertGroup(ctx context.Context, serverID int64, clientID, groupID, senderID string, msgType int32, text, attachment string, serverTime int64) error {
	return r.db.WithContext(ctx).Create(&GroupMessageModel{
		MessageServerID: serverID, MessageClientID: clientID, GroupID: groupID,
		SenderID: senderID, MsgType: msgType, Text: text, Attachment: attachment,
		ServerTime: serverTime, Status: 1,
	}).Error
}

// GetReceiverID returns the receiver of a message (lightweight lookup).
func (r *MessageRepo) GetReceiverID(ctx context.Context, serverID int64) (string, error) {
	var m MessageModel
	err := r.db.WithContext(ctx).Select("receiver_id").Where("message_server_id = ?", serverID).First(&m).Error
	if err != nil {
		return "", err
	}
	return m.ReceiverID, nil
}

// MarkRecalled atomically updates the message status to RECALLED if the
// sender matches and the message is within the 2-minute window and not
// already recalled. Returns true if the update was applied.
func (r *MessageRepo) MarkRecalled(ctx context.Context, serverID int64, senderID string, serverTime int64) (bool, error) {
	cutoff := serverTime - 2*60*1000 // 2 minutes ago
	result := r.db.WithContext(ctx).Model(&MessageModel{}).
		Where("message_server_id = ? AND sender_id = ? AND server_time >= ? AND status != ?",
			serverID, senderID, cutoff, int32(msgpb.MessageStatus_MESSAGE_STATUS_RECALLED)).
		Update("status", int32(msgpb.MessageStatus_MESSAGE_STATUS_RECALLED))
	return result.RowsAffected > 0, result.Error
}

// Insert inserts a message with status SENT.
func (r *MessageRepo) Insert(ctx context.Context, m *biz.Message) error {
	model := &MessageModel{
		MessageServerID: m.ServerID,
		MessageClientID: m.ClientID,
		SenderID:        m.SenderID,
		ReceiverID:      m.ReceiverID,
		ConvType:        m.ConvType,
		MsgType:         m.MsgType,
		Text:            m.Text,
		Attachment:      string(m.Attachment),
		ServerTime:      m.ServerTime,
		CreateTime:      m.CreateTime,
		Status:          int32(msgpb.MessageStatus_MESSAGE_STATUS_SENT),
	}
	return r.db.WithContext(ctx).Clauses(clause.OnConflict{DoNothing: true}).Create(model).Error
}

// MarkDelivered updates the message status to DELIVERED.
func (r *MessageRepo) MarkDelivered(ctx context.Context, serverID int64) error {
	return r.db.WithContext(ctx).Model(&MessageModel{}).
		Where("message_server_id = ?", serverID).
		Update("status", int32(msgpb.MessageStatus_MESSAGE_STATUS_DELIVERED)).Error
}

// MarkRead updates the message status to READ.
func (r *MessageRepo) MarkRead(ctx context.Context, serverID int64) error {
	return r.db.WithContext(ctx).Model(&MessageModel{}).
		Where("message_server_id = ?", serverID).
		Updates(map[string]interface{}{
			"is_remote_read": true,
			"status":         int32(msgpb.MessageStatus_MESSAGE_STATUS_READ),
		}).Error
}

// PullSince returns messages for a user with server ID greater than sinceID.
func (r *MessageRepo) PullSince(ctx context.Context, userID string, sinceID int64, limit int32) ([]biz.Message, error) {
	var models []MessageModel
	err := r.db.WithContext(ctx).
		Where("receiver_id = ? AND message_server_id > ? AND is_deleted = false", userID, sinceID).
		Order("message_server_id ASC").
		Limit(int(limit)).
		Find(&models).Error
	if err != nil {
		return nil, err
	}
	msgs := make([]biz.Message, len(models))
	for i, m := range models {
		msgs[i] = biz.Message{
			ServerID:   m.MessageServerID,
			ClientID:   m.MessageClientID,
			SenderID:   m.SenderID,
			ReceiverID: m.ReceiverID,
			ConvType:   m.ConvType,
			MsgType:    m.MsgType,
			Text:       m.Text,
			Attachment: []byte(m.Attachment),
			ServerTime: m.ServerTime,
			CreateTime: m.CreateTime,
			IsDeleted:  m.IsDeleted,
			IsRead:     m.IsRemoteRead,
			Status:     m.Status,
		}
	}
	return msgs, nil
}
