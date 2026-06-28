package data

import "time"

// RoomModel is the GORM model for chat rooms.
type RoomModel struct {
	ID          int64     `gorm:"primaryKey;autoIncrement"`
	RoomID      string    `gorm:"column:room_id;uniqueIndex;size:64;not null"`
	Name        string    `gorm:"column:name;size:128;not null"`
	OwnerID     string    `gorm:"column:owner_id;size:64;not null"`
	MemberCount int32     `gorm:"column:member_count;default:0"`
	CreatedAt   time.Time `gorm:"column:created_at;autoCreateTime"`
}

func (RoomModel) TableName() string { return "rooms" }

// RoomMemberModel is the GORM model for room membership.
type RoomMemberModel struct {
	ID       int64     `gorm:"primaryKey;autoIncrement"`
	RoomID   string    `gorm:"column:room_id;size:64;not null;uniqueIndex:idx_room_user"`
	UserID   string    `gorm:"column:user_id;size:64;not null;uniqueIndex:idx_room_user"`
	JoinedAt time.Time `gorm:"column:joined_at;autoCreateTime"`
}

func (RoomMemberModel) TableName() string { return "room_members" }

// ChatroomMessageModel reads from chatroom_messages table (shared with message service).
type ChatroomMessageModel struct {
	ID         int64  `gorm:"primaryKey;autoIncrement"`
	ServerID   string `gorm:"column:server_id;size:64;not null"`
	RoomID     string `gorm:"column:room_id;size:64;not null;index"`
	SenderID   string `gorm:"column:sender_id;size:64;not null"`
	MsgType    int32  `gorm:"column:msg_type"`
	Text       string `gorm:"column:text"`
	Attachment []byte `gorm:"column:attachment"`
	CreateTime int64  `gorm:"column:create_time"`
}

func (ChatroomMessageModel) TableName() string { return "chatroom_messages" }
