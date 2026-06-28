package data

import "time"

// UserModel is the GORM model for the users table.
type UserModel struct {
	ID           int64     `gorm:"primaryKey;autoIncrement"`
	UserID       string    `gorm:"column:user_id;uniqueIndex;size:64;not null"`
	Username     string    `gorm:"column:username;uniqueIndex;size:64;not null"`
	PasswordHash string    `gorm:"column:password_hash;size:256;not null"`
	Nickname     string    `gorm:"column:nickname;size:128"`
	Avatar       string    `gorm:"column:avatar;size:512"`
	CreatedAt    time.Time `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt    time.Time `gorm:"column:updated_at;autoUpdateTime"`
}

func (UserModel) TableName() string { return "users" }
