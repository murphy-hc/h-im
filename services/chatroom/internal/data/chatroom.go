package data

import (
	"context"
	"fmt"

	"github.com/murphy-hc/h-im/services/chatroom/internal/biz"
	"gorm.io/gorm"
)

var _ biz.ChatroomRepo = (*chatroomRepo)(nil)

type chatroomRepo struct {
	data *Data
}

func NewChatroomRepo(data *Data) biz.ChatroomRepo {
	return &chatroomRepo{data: data}
}

func (r *chatroomRepo) CreateRoom(ctx context.Context, roomID, name, ownerID string) error {
	return r.data.DB.WithContext(ctx).Create(&RoomModel{RoomID: roomID, Name: name, OwnerID: ownerID, MemberCount: 1}).Error
}

func (r *chatroomRepo) FindByID(ctx context.Context, roomID string) (*biz.Room, error) {
	var m RoomModel
	if err := r.data.DB.WithContext(ctx).Where("room_id = ?", roomID).First(&m).Error; err != nil {
		return nil, err
	}
	return &biz.Room{RoomID: m.RoomID, Name: m.Name, OwnerID: m.OwnerID, MemberCount: m.MemberCount}, nil
}

func (r *chatroomRepo) JoinRoom(ctx context.Context, roomID, userID string) error {
	return r.data.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&RoomMemberModel{RoomID: roomID, UserID: userID}).Error; err != nil {
			return err
		}
		return tx.Model(&RoomModel{}).Where("room_id = ?", roomID).UpdateColumn("member_count", gorm.Expr("member_count + 1")).Error
	})
}

func (r *chatroomRepo) LeaveRoom(ctx context.Context, roomID, userID string) error {
	return r.data.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		res := tx.Where("room_id = ? AND user_id = ?", roomID, userID).Delete(&RoomMemberModel{})
		if res.RowsAffected == 0 {
			return fmt.Errorf("not a member")
		}
		return tx.Model(&RoomModel{}).Where("room_id = ?", roomID).UpdateColumn("member_count", gorm.Expr("member_count - 1")).Error
	})
}

func (r *chatroomRepo) GetMembers(ctx context.Context, roomID string) ([]string, error) {
	var members []RoomMemberModel
	if err := r.data.DB.WithContext(ctx).Where("room_id = ?", roomID).Find(&members).Error; err != nil {
		return nil, err
	}
	ids := make([]string, len(members))
	for i, m := range members {
		ids[i] = m.UserID
	}
	return ids, nil
}
