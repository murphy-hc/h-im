package data

import "time"

// MediaModel is the GORM model for the media table.
type MediaModel struct {
	ID          int64     `gorm:"primaryKey;autoIncrement"`
	MediaID     string    `gorm:"column:media_id;uniqueIndex;size:64;not null"`
	UserID      string    `gorm:"column:user_id;size:64;not null;index"`
	MediaType   int32     `gorm:"column:media_type;not null"`
	URL         string    `gorm:"column:url;size:1024;not null"`
	ThumbURL    string    `gorm:"column:thumb_url;size:1024"`
	FileName    string    `gorm:"column:file_name;size:256"`
	MimeType    string    `gorm:"column:mime_type;size:128"`
	Size        int64     `gorm:"column:size;default:0"`
	CreatedAt   time.Time `gorm:"column:created_at;autoCreateTime"`
}

func (MediaModel) TableName() string { return "medias" }
