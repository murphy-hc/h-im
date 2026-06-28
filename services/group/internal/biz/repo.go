package biz

import "context"

// Group is the domain entity for a group.
type Group struct {
	GroupID      string
	Name         string
	OwnerID      string
	Announcement string
	MemberCount  int32
}

// GroupMember is a group member with role.
type GroupMember struct {
	UserID string
	Role   int32
}

// GroupRepo defines the group repository interface.
type GroupRepo interface {
	Create(ctx context.Context, g *Group) error
	Update(ctx context.Context, g *Group) error
	Delete(ctx context.Context, groupID string) error
	FindByID(ctx context.Context, groupID string) (*Group, error)
	Join(ctx context.Context, groupID, userID string, role int32) error
	Leave(ctx context.Context, groupID, userID string) error
	KickMember(ctx context.Context, groupID, userID string) error
	SetRole(ctx context.Context, groupID, userID string, role int32) error
	GetMembers(ctx context.Context, groupID string, offset, limit int32) ([]GroupMember, error)
	MuteMember(ctx context.Context, groupID, userID string, seconds int32) error
	UnmuteMember(ctx context.Context, groupID, userID string) error
	IsMuted(ctx context.Context, groupID, userID string) (bool, error)
}
