package data

import (
	"context"
	"fmt"

	"github.com/murphy-hc/h-im/services/contact/internal/biz"
	"gorm.io/gorm"
)

var _ biz.ContactRepo = (*contactRepo)(nil)

type contactRepo struct {
	data *Data
}

func NewContactRepo(data *Data) biz.ContactRepo {
	return &contactRepo{data: data}
}

func (r *contactRepo) SendRequest(ctx context.Context, reqID, fromUser, toUser, msg string) error {
	return r.data.DB.WithContext(ctx).Create(&FriendRequestModel{
		RequestID: reqID, FromUser: fromUser, ToUser: toUser, Message: msg, Status: 0,
	}).Error
}

func (r *contactRepo) RespondRequest(ctx context.Context, reqID string, accept bool) error {
	return r.data.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var req FriendRequestModel
		if err := tx.Where("request_id = ? AND status = 0", reqID).First(&req).Error; err != nil {
			return fmt.Errorf("request not found")
		}
		status := int32(2) // rejected
		if accept {
			status = 1 // accepted
			if err := tx.Create(&FriendModel{UserID: req.FromUser, FriendID: req.ToUser, Status: 1}).Error; err != nil {
				return err
			}
			if err := tx.Create(&FriendModel{UserID: req.ToUser, FriendID: req.FromUser, Status: 1}).Error; err != nil {
				return err
			}
		}
		return tx.Model(&req).Update("status", status).Error
	})
}

func (r *contactRepo) FindPendingRequest(ctx context.Context, fromUser, toUser string) (*biz.FriendRequest, error) {
	var m FriendRequestModel
	if err := r.data.DB.WithContext(ctx).Where("from_user = ? AND to_user = ? AND status = 0", fromUser, toUser).First(&m).Error; err != nil {
		return nil, err
	}
	return &biz.FriendRequest{RequestID: m.RequestID, FromUser: m.FromUser, ToUser: m.ToUser, Message: m.Message, Status: m.Status}, nil
}

func (r *contactRepo) RemoveFriend(ctx context.Context, userID, friendID string) error {
	return r.data.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		tx.Where("user_id = ? AND friend_id = ?", userID, friendID).Delete(&FriendModel{})
		tx.Where("user_id = ? AND friend_id = ?", friendID, userID).Delete(&FriendModel{})
		return nil
	})
}

func (r *contactRepo) BlockUser(ctx context.Context, userID, blockID string) error {
	return r.data.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		tx.Where("user_id = ? AND friend_id = ?", userID, blockID).Delete(&FriendModel{})
		tx.Where("user_id = ? AND friend_id = ?", blockID, userID).Delete(&FriendModel{})
		return tx.Create(&FriendModel{UserID: userID, FriendID: blockID, Status: 2}).Error
	})
}

func (r *contactRepo) UnblockUser(ctx context.Context, userID, blockID string) error {
	return r.data.DB.WithContext(ctx).Where("user_id = ? AND friend_id = ? AND status = 2", userID, blockID).Delete(&FriendModel{}).Error
}

func (r *contactRepo) GetFriends(ctx context.Context, userID string, offset, limit int32) ([]biz.FriendInfo, error) {
	var models []FriendModel
	if err := r.data.DB.WithContext(ctx).Where("user_id = ? AND status = 1", userID).Offset(int(offset)).Limit(int(limit)).Find(&models).Error; err != nil {
		return nil, err
	}
	friends := make([]biz.FriendInfo, len(models))
	for i, m := range models {
		friends[i] = biz.FriendInfo{UserID: m.FriendID, Status: m.Status}
	}
	return friends, nil
}

func (r *contactRepo) GetRequests(ctx context.Context, userID string) ([]biz.FriendRequest, error) {
	var models []FriendRequestModel
	if err := r.data.DB.WithContext(ctx).Where("to_user = ? AND status = 0", userID).Order("created_at DESC").Find(&models).Error; err != nil {
		return nil, err
	}
	reqs := make([]biz.FriendRequest, len(models))
	for i, m := range models {
		reqs[i] = biz.FriendRequest{RequestID: m.RequestID, FromUser: m.FromUser, ToUser: m.ToUser, Message: m.Message, Status: m.Status}
	}
	return reqs, nil
}
