package data

import "time"

// PushDeviceModel is the GORM model for push device registrations.
type PushDeviceModel struct {
	ID           int64     `gorm:"primaryKey;autoIncrement"`
	UserID       string    `gorm:"column:user_id;size:64;not null;index"`
	DeviceID     string    `gorm:"column:device_id;uniqueIndex;size:64;not null"`
	DeviceToken  string    `gorm:"column:device_token;size:512;not null"`
	Platform     int32     `gorm:"column:platform;default:0"`
	CreatedAt    time.Time `gorm:"column:created_at;autoCreateTime"`
}

func (PushDeviceModel) TableName() string { return "push_devices" }
