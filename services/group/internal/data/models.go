package data

import "time"

// GroupModel is the GORM model for groups.
type GroupModel struct {
	ID           int64     `gorm:"primaryKey;autoIncrement"`
	GroupID      string    `gorm:"column:group_id;uniqueIndex;size:64;not null"`
	Name         string    `gorm:"column:name;size:128;not null"`
	OwnerID      string    `gorm:"column:owner_id;size:64;not null"`
	Announcement string    `gorm:"column:announcement;type:text"`
	MemberCount  int32     `gorm:"column:member_count;default:0"`
	Status       int32     `gorm:"column:status;default:0"`
	CreatedAt    time.Time `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt    time.Time `gorm:"column:updated_at;autoUpdateTime"`
}

func (GroupModel) TableName() string { return "groups" }

// GroupMemberModel is the GORM model for group membership.
type GroupMemberModel struct {
	ID       int64     `gorm:"primaryKey;autoIncrement"`
	GroupID  string    `gorm:"column:group_id;size:64;not null;uniqueIndex:idx_group_user"`
	UserID   string    `gorm:"column:user_id;size:64;not null;uniqueIndex:idx_group_user"`
	Role     int32     `gorm:"column:role;default:2"`
	JoinedAt time.Time `gorm:"column:joined_at;autoCreateTime"`
}

func (GroupMemberModel) TableName() string { return "group_members" }
