package service

import (
	"context"
	"testing"

	pb "github.com/murphy-hc/h-im/gen/go/him/group/v1"
	"github.com/murphy-hc/h-im/services/group/internal/biz"
)

type e2eGroupRepo struct {
	groups  map[string]*biz.Group
	members map[string]map[string]int32
	mutes   map[string]bool
}

func newE2EGroupRepo() *e2eGroupRepo {
	return &e2eGroupRepo{
		groups:  make(map[string]*biz.Group),
		members: make(map[string]map[string]int32),
		mutes:   make(map[string]bool),
	}
}

func (r *e2eGroupRepo) Create(_ context.Context, g *biz.Group) error {
	r.groups[g.GroupID] = g
	r.members[g.GroupID] = make(map[string]int32)
	return nil
}
func (r *e2eGroupRepo) Update(_ context.Context, _ *biz.Group) error { return nil }
func (r *e2eGroupRepo) Delete(_ context.Context, groupID string) error {
	delete(r.groups, groupID); return nil
}
func (r *e2eGroupRepo) FindByID(_ context.Context, groupID string) (*biz.Group, error) {
	g, ok := r.groups[groupID]
	if !ok {
		return nil, context.DeadlineExceeded
	}
	return g, nil
}
func (r *e2eGroupRepo) Join(_ context.Context, groupID, userID string, role int32) error {
	r.members[groupID][userID] = role; return nil
}
func (r *e2eGroupRepo) Leave(_ context.Context, groupID, userID string) error {
	delete(r.members[groupID], userID); return nil
}
func (r *e2eGroupRepo) KickMember(_ context.Context, groupID, userID string) error {
	delete(r.members[groupID], userID); return nil
}
func (r *e2eGroupRepo) SetRole(_ context.Context, groupID, userID string, role int32) error {
	r.members[groupID][userID] = role; return nil
}
func (r *e2eGroupRepo) GetMembers(_ context.Context, groupID string, _, _ int32) ([]biz.GroupMember, error) {
	var out []biz.GroupMember
	for id, role := range r.members[groupID] {
		out = append(out, biz.GroupMember{UserID: id, Role: role})
	}
	return out, nil
}
func (r *e2eGroupRepo) MuteMember(_ context.Context, groupID, userID string, _ int32) error {
	r.mutes[groupID+":"+userID] = true; return nil
}
func (r *e2eGroupRepo) UnmuteMember(_ context.Context, groupID, userID string) error {
	delete(r.mutes, groupID+":"+userID); return nil
}
func (r *e2eGroupRepo) IsMuted(_ context.Context, groupID, userID string) (bool, error) {
	return r.mutes[groupID+":"+userID], nil
}

func newE2EGroupService() *GroupService {
	return NewGroupService(biz.NewGroupUseCase(newE2EGroupRepo()))
}

func TestE2E_CreateJoinMuteKickDismiss(t *testing.T) {
	svc := newE2EGroupService()
	ctx := context.Background()

	// 1. Create group
	createResp, err := svc.CreateGroup(ctx, &pb.CreateGroupRequest{Name: "E2E Group", OwnerId: "owner"})
	if err != nil {
		t.Fatalf("CreateGroup: %v", err)
	}
	gid := createResp.GroupId

	// 2. Join
	_, err = svc.JoinGroup(ctx, &pb.JoinGroupRequest{GroupId: gid, UserId: "member-1"})
	if err != nil {
		t.Fatalf("JoinGroup: %v", err)
	}

	// 3. Get members
	members, err := svc.GetGroupMembers(ctx, &pb.GetGroupMembersRequest{GroupId: gid})
	if err != nil {
		t.Fatalf("GetGroupMembers: %v", err)
	}
	if len(members.Members) != 2 {
		t.Fatalf("expected 2 members, got %d", len(members.Members))
	}

	// 4. Mute
	_, err = svc.MuteMember(ctx, &pb.MuteMemberRequest{GroupId: gid, UserId: "member-1", OperatorId: "owner", DurationSeconds: 3600})
	if err != nil {
		t.Fatalf("MuteMember: %v", err)
	}

	// 5. Kick
	_, err = svc.KickMember(ctx, &pb.KickMemberRequest{GroupId: gid, UserId: "member-1", OperatorId: "owner"})
	if err != nil {
		t.Fatalf("KickMember: %v", err)
	}

	// 6. Dismiss
	_, err = svc.DismissGroup(ctx, &pb.DismissGroupRequest{GroupId: gid, OwnerId: "owner"})
	if err != nil {
		t.Fatalf("DismissGroup: %v", err)
	}
}
