package data

import (
	"context"
	"fmt"
	"time"

	"github.com/murphy-hc/h-im/services/group/internal/biz"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

var _ biz.GroupRepo = (*groupRepo)(nil)

type groupRepo struct {
	data *Data
}

func NewGroupRepo(data *Data) biz.GroupRepo {
	return &groupRepo{data: data}
}

func muteKey(groupID, userID string) string { return fmt.Sprintf("group:mute:%s:%s", groupID, userID) }

func (r *groupRepo) Create(ctx context.Context, g *biz.Group) error {
	return r.data.DB.WithContext(ctx).Create(&GroupModel{
		GroupID: g.GroupID, Name: g.Name, OwnerID: g.OwnerID,
		Announcement: g.Announcement, MemberCount: g.MemberCount,
	}).Error
}

func (r *groupRepo) Update(ctx context.Context, g *biz.Group) error {
	return r.data.DB.WithContext(ctx).Model(&GroupModel{}).Where("group_id = ?", g.GroupID).
		Updates(map[string]any{"name": g.Name, "announcement": g.Announcement}).Error
}

func (r *groupRepo) Delete(ctx context.Context, groupID string) error {
	return r.data.DB.WithContext(ctx).Where("group_id = ?", groupID).Delete(&GroupModel{}).Error
}

func (r *groupRepo) FindByID(ctx context.Context, groupID string) (*biz.Group, error) {
	var m GroupModel
	if err := r.data.DB.WithContext(ctx).Where("group_id = ?", groupID).First(&m).Error; err != nil {
		return nil, err
	}
	return &biz.Group{GroupID: m.GroupID, Name: m.Name, OwnerID: m.OwnerID, Announcement: m.Announcement, MemberCount: m.MemberCount}, nil
}

func (r *groupRepo) Join(ctx context.Context, groupID, userID string, role int32) error {
	return r.data.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&GroupMemberModel{GroupID: groupID, UserID: userID, Role: role}).Error; err != nil {
			return err
		}
		return tx.Model(&GroupModel{}).Where("group_id = ?", groupID).
			UpdateColumn("member_count", gorm.Expr("member_count + 1")).Error
	})
}

func (r *groupRepo) Leave(ctx context.Context, groupID, userID string) error {
	return r.data.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		res := tx.Where("group_id = ? AND user_id = ?", groupID, userID).Delete(&GroupMemberModel{})
		if res.RowsAffected == 0 {
			return fmt.Errorf("not a member")
		}
		return tx.Model(&GroupModel{}).Where("group_id = ?", groupID).
			UpdateColumn("member_count", gorm.Expr("member_count - 1")).Error
	})
}

func (r *groupRepo) KickMember(ctx context.Context, groupID, userID string) error {
	return r.data.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		res := tx.Where("group_id = ? AND user_id = ?", groupID, userID).Delete(&GroupMemberModel{})
		if res.RowsAffected == 0 {
			return fmt.Errorf("user not in group")
		}
		return tx.Model(&GroupModel{}).Where("group_id = ?", groupID).
			UpdateColumn("member_count", gorm.Expr("member_count - 1")).Error
	})
}

func (r *groupRepo) SetRole(ctx context.Context, groupID, userID string, role int32) error {
	return r.data.DB.WithContext(ctx).Model(&GroupMemberModel{}).
		Where("group_id = ? AND user_id = ?", groupID, userID).Update("role", role).Error
}

func (r *groupRepo) GetMembers(ctx context.Context, groupID string, offset, limit int32) ([]biz.GroupMember, error) {
	var models []GroupMemberModel
	err := r.data.DB.WithContext(ctx).Where("group_id = ?", groupID).
		Offset(int(offset)).Limit(int(limit)).Find(&models).Error
	if err != nil {
		return nil, err
	}
	members := make([]biz.GroupMember, len(models))
	for i, m := range models {
		members[i] = biz.GroupMember{UserID: m.UserID, Role: m.Role}
	}
	return members, nil
}

func (r *groupRepo) MuteMember(ctx context.Context, groupID, userID string, seconds int32) error {
	return r.data.RDB.Set(ctx, muteKey(groupID, userID), 1, time.Duration(seconds)*time.Second).Err()
}

func (r *groupRepo) UnmuteMember(ctx context.Context, groupID, userID string) error {
	return r.data.RDB.Del(ctx, muteKey(groupID, userID)).Err()
}

func (r *groupRepo) IsMuted(ctx context.Context, groupID, userID string) (bool, error) {
	n, err := r.data.RDB.Exists(ctx, muteKey(groupID, userID)).Result()
	return n > 0, err
}

var _ redis.Cmdable = (*redis.Client)(nil)
