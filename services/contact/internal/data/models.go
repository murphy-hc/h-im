package data

import "time"

// FriendModel is the GORM model for friend relationships.
type FriendModel struct {
	ID        int64     `gorm:"primaryKey;autoIncrement"`
	UserID    string    `gorm:"column:user_id;size:64;not null;index:idx_user"`
	FriendID  string    `gorm:"column:friend_id;size:64;not null"`
	Status    int32     `gorm:"column:status;default:0"`
	CreatedAt time.Time `gorm:"column:created_at;autoCreateTime"`
}

func (FriendModel) TableName() string { return "friends" }

// FriendRequestModel is the GORM model for friend requests.
type FriendRequestModel struct {
	ID        int64     `gorm:"primaryKey;autoIncrement"`
	RequestID string    `gorm:"column:request_id;uniqueIndex;size:64;not null"`
	FromUser  string    `gorm:"column:from_user;size:64;not null"`
	ToUser    string    `gorm:"column:to_user;size:64;not null;index:idx_to_user"`
	Message   string    `gorm:"column:message;size:256"`
	Status    int32     `gorm:"column:status;default:0"`
	CreatedAt time.Time `gorm:"column:created_at;autoCreateTime"`
}

func (FriendRequestModel) TableName() string { return "friend_requests" }
