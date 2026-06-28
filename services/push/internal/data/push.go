package data

import (
	"context"

	"github.com/murphy-hc/h-im/services/push/internal/biz"
	"gorm.io/gorm/clause"
)

var _ biz.PushRepo = (*pushRepo)(nil)

type pushRepo struct {
	data *Data
}

func NewPushRepo(data *Data) biz.PushRepo { return &pushRepo{data: data} }

func (r *pushRepo) RegisterDevice(ctx context.Context, userID string, info *biz.DeviceInfo) error {
	return r.data.DB.WithContext(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "device_id"}},
		DoUpdates: clause.AssignmentColumns([]string{"device_token", "platform"}),
	}).Create(&PushDeviceModel{
		UserID: userID, DeviceID: info.DeviceID, DeviceToken: info.DeviceToken, Platform: info.Platform,
	}).Error
}

func (r *pushRepo) UnregisterDevice(ctx context.Context, deviceID string) error {
	return r.data.DB.WithContext(ctx).Where("device_id = ?", deviceID).Delete(&PushDeviceModel{}).Error
}

func (r *pushRepo) FindDevicesByUser(ctx context.Context, userID string) ([]biz.DeviceInfo, error) {
	var models []PushDeviceModel
	if err := r.data.DB.WithContext(ctx).Where("user_id = ?", userID).Find(&models).Error; err != nil {
		return nil, err
	}
	devices := make([]biz.DeviceInfo, len(models))
	for i, m := range models {
		devices[i] = biz.DeviceInfo{DeviceID: m.DeviceID, DeviceToken: m.DeviceToken, Platform: m.Platform}
	}
	return devices, nil
}
