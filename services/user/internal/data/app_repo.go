package data

import (
	"context"
	"fmt"
	"time"

	"github.com/murphy-hc/h-im/services/user/internal/biz"
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
func (a *AppRepo) FindByAppID(ctx context.Context, appID string) (*biz.App, error) {
	var model AppModel
	err := a.db.WithContext(ctx).Where("app_id = ? AND enabled = true", appID).First(&model).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("app not found: %s", appID)
		}
		return nil, fmt.Errorf("query app: %w", err)
	}
	return &biz.App{AppID: model.AppID, AppSecret: model.AppSecret, AppName: model.AppName, Enabled: model.Enabled}, nil
}
