package biz

import (
	"context"
	"fmt"

	"github.com/rs/xid"
)

// GroupUseCase handles group business logic.
type GroupUseCase struct {
	repo GroupRepo
}

// NewGroupUseCase creates a GroupUseCase.
func NewGroupUseCase(repo GroupRepo) *GroupUseCase {
	return &GroupUseCase{repo: repo}
}

func (uc *GroupUseCase) CreateGroup(ctx context.Context, name, ownerID string) (*Group, error) {
	g := &Group{GroupID: xid.New().String(), Name: name, OwnerID: ownerID, MemberCount: 1}
	if err := uc.repo.Create(ctx, g); err != nil {
		return nil, fmt.Errorf("create: %w", err)
	}
	if err := uc.repo.Join(ctx, g.GroupID, ownerID, 0); err != nil {
		return nil, fmt.Errorf("join owner: %w", err)
	}
	return g, nil
}

func (uc *GroupUseCase) UpdateGroup(ctx context.Context, groupID, name, announcement string) error {
	g, err := uc.repo.FindByID(ctx, groupID)
	if err != nil {
		return fmt.Errorf("not found: %w", err)
	}
	if name != "" {
		g.Name = name
	}
	g.Announcement = announcement
	return uc.repo.Update(ctx, g)
}

func (uc *GroupUseCase) DismissGroup(ctx context.Context, groupID, ownerID string) error {
	g, err := uc.repo.FindByID(ctx, groupID)
	if err != nil {
		return fmt.Errorf("not found: %w", err)
	}
	if g.OwnerID != ownerID {
		return fmt.Errorf("only owner can dismiss")
	}
	return uc.repo.Delete(ctx, groupID)
}

func (uc *GroupUseCase) JoinGroup(ctx context.Context, groupID, userID string) error {
	if _, err := uc.repo.FindByID(ctx, groupID); err != nil {
		return fmt.Errorf("not found: %w", err)
	}
	return uc.repo.Join(ctx, groupID, userID, 2)
}

func (uc *GroupUseCase) LeaveGroup(ctx context.Context, groupID, userID string) error {
	return uc.repo.Leave(ctx, groupID, userID)
}

func (uc *GroupUseCase) GetGroupInfo(ctx context.Context, groupID string) (*Group, error) {
	return uc.repo.FindByID(ctx, groupID)
}

func (uc *GroupUseCase) GetGroupMembers(ctx context.Context, groupID string, offset, limit int32) ([]GroupMember, error) {
	return uc.repo.GetMembers(ctx, groupID, offset, limit)
}

func (uc *GroupUseCase) SetMemberRole(ctx context.Context, groupID, userID, operatorID string, role int32) error {
	g, _ := uc.repo.FindByID(ctx, groupID)
	if g != nil && g.OwnerID != operatorID {
		return fmt.Errorf("only owner can set role")
	}
	return uc.repo.SetRole(ctx, groupID, userID, role)
}

func (uc *GroupUseCase) KickMember(ctx context.Context, groupID, userID, operatorID string) error {
	g, _ := uc.repo.FindByID(ctx, groupID)
	if g != nil && g.OwnerID != operatorID {
		return fmt.Errorf("only owner can kick")
	}
	return uc.repo.KickMember(ctx, groupID, userID)
}

func (uc *GroupUseCase) MuteMember(ctx context.Context, groupID, userID, operatorID string, seconds int32) error {
	g, _ := uc.repo.FindByID(ctx, groupID)
	if g != nil && g.OwnerID != operatorID {
		return fmt.Errorf("only owner can mute")
	}
	return uc.repo.MuteMember(ctx, groupID, userID, seconds)
}

func (uc *GroupUseCase) UnmuteMember(ctx context.Context, groupID, userID string) error {
	return uc.repo.UnmuteMember(ctx, groupID, userID)
}

func (uc *GroupUseCase) IsMuted(ctx context.Context, groupID, userID string) (bool, error) {
	return uc.repo.IsMuted(ctx, groupID, userID)
}
