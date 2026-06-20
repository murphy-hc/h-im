package data

import (
	"context"
	"fmt"
	"time"

	"gorm.io/gorm"
)

// AppModel is the GORM model for the apps table.
type AppModel struct {
	ID        int64     `gorm:"primaryKey;autoIncrement"`
	AppID     string    `gorm:"column:app_id;uniqueIndex;size:64;not null"`
	AppSecret string    `gorm:"column:app_secret;size:256;not null"`
	AppName   string    `gorm:"column:app_name;size:128"`
	Enabled   bool      `gorm:"column:enabled;default:true"`
	CreatedAt time.Time `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt time.Time `gorm:"column:updated_at;autoUpdateTime"`
}

func (AppModel) TableName() string { return "apps" }

// AppRepo provides read access to the apps table.
type AppRepo struct {
	db *gorm.DB
}

// NewAppRepo creates an AppRepo.
func NewAppRepo(data *Data) *AppRepo {
	return &AppRepo{db: data.DB}
}

// FindByAppID looks up an enabled app by its app_id.
func (r *AppRepo) FindByAppID(ctx context.Context, appID string) (*AppModel, error) {
	var app AppModel
	err := r.db.WithContext(ctx).Where("app_id = ? AND enabled = true", appID).First(&app).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("app not found: %s", appID)
		}
		return nil, fmt.Errorf("query app: %w", err)
	}
	return &app, nil
}
