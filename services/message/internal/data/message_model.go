package data

import "time"

// MessageModel is the GORM model for private messages.
type MessageModel struct {
	MessageServerID int64  `gorm:"column:message_server_id;primaryKey"`
	MessageClientID string `gorm:"column:message_client_id;uniqueIndex;size:64;not null"`
	SenderID        string `gorm:"column:sender_id;size:64;not null;index:idx_sender"`
	ReceiverID      string `gorm:"column:receiver_id;size:64;not null;index:idx_receiver"`
	ConvType        int32  `gorm:"column:conv_type;not null"`
	MsgType         int32  `gorm:"column:msg_type;not null;default:0"`
	MsgSubType      int32  `gorm:"column:msg_sub_type;default:0"`
	Text            string `gorm:"column:text"`
	Attachment      string `gorm:"column:attachment;type:jsonb"`
	ServerTime      int64  `gorm:"column:server_time;not null"`
	CreateTime      int64  `gorm:"column:create_time;not null;default:0"`
	IsDeleted       bool   `gorm:"column:is_deleted;default:false"`
	IsRemoteRead    bool   `gorm:"column:is_remote_read;default:false"`
	Status          int32  `gorm:"column:status;default:0"`
	CreatedAt       time.Time `gorm:"column:created_at;autoCreateTime"`
}

func (MessageModel) TableName() string { return "private_messages" }

// ChatroomMessageModel is the GORM model for chatroom messages.
type ChatroomMessageModel struct {
	MessageServerID int64     `gorm:"column:message_server_id;primaryKey"`
	MessageClientID string    `gorm:"column:message_client_id;uniqueIndex;size:64;not null"`
	RoomID          string    `gorm:"column:room_id;size:64;not null;index:idx_room"`
	SenderID        string    `gorm:"column:sender_id;size:64;not null;index:idx_sender"`
	MsgType         int32     `gorm:"column:msg_type;not null;default:0"`
	Text            string    `gorm:"column:text"`
	Attachment      string    `gorm:"column:attachment;type:jsonb"`
	Priority        int32     `gorm:"column:priority;default:0"`
	ServerTime      int64     `gorm:"column:server_time;not null"`
	Status          int32     `gorm:"column:status;default:0"`
	CreatedAt       time.Time `gorm:"column:created_at;autoCreateTime"`
}

func (ChatroomMessageModel) TableName() string { return "chatroom_messages" }

// GroupMessageModel is the GORM model for group messages.
type GroupMessageModel struct {
	MessageServerID int64     `gorm:"column:message_server_id;primaryKey"`
	MessageClientID string    `gorm:"column:message_client_id;uniqueIndex;size:64;not null"`
	GroupID         string    `gorm:"column:group_id;size:64;not null;index:idx_group"`
	SenderID        string    `gorm:"column:sender_id;size:64;not null;index:idx_sender"`
	MsgType         int32     `gorm:"column:msg_type;not null;default:0"`
	Text            string    `gorm:"column:text"`
	Attachment      string    `gorm:"column:attachment;type:jsonb"`
	ServerTime      int64     `gorm:"column:server_time;not null"`
	Status          int32     `gorm:"column:status;default:0"`
	CreatedAt       time.Time `gorm:"column:created_at;autoCreateTime"`
}

func (GroupMessageModel) TableName() string { return "group_messages" }
