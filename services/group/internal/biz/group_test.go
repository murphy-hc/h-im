package biz

import (
	"context"
	"testing"
)

type mockGroupRepo struct {
	groups  map[string]*Group
	members map[string]map[string]int32
	mutes   map[string]bool
}

func newMockGroupRepo() *mockGroupRepo {
	return &mockGroupRepo{
		groups:  make(map[string]*Group),
		members: make(map[string]map[string]int32),
		mutes:   make(map[string]bool),
	}
}

func (m *mockGroupRepo) Create(ctx context.Context, g *Group) error {
	m.groups[g.GroupID] = g
	m.members[g.GroupID] = make(map[string]int32)
	return nil
}
func (m *mockGroupRepo) Update(ctx context.Context, g *Group) error { return nil }
func (m *mockGroupRepo) Delete(ctx context.Context, groupID string) error {
	delete(m.groups, groupID)
	return nil
}
func (m *mockGroupRepo) FindByID(ctx context.Context, groupID string) (*Group, error) {
	g, ok := m.groups[groupID]
	if !ok {
		return nil, context.DeadlineExceeded
	}
	return g, nil
}
func (m *mockGroupRepo) Join(ctx context.Context, groupID, userID string, role int32) error {
	m.members[groupID][userID] = role
	return nil
}
func (m *mockGroupRepo) Leave(ctx context.Context, groupID, userID string) error {
	delete(m.members[groupID], userID)
	return nil
}
func (m *mockGroupRepo) KickMember(ctx context.Context, groupID, userID string) error {
	delete(m.members[groupID], userID)
	return nil
}
func (m *mockGroupRepo) SetRole(ctx context.Context, groupID, userID string, role int32) error {
	m.members[groupID][userID] = role
	return nil
}
func (m *mockGroupRepo) GetMembers(ctx context.Context, groupID string, offset, limit int32) ([]GroupMember, error) {
	var out []GroupMember
	for id, role := range m.members[groupID] {
		out = append(out, GroupMember{UserID: id, Role: role})
	}
	return out, nil
}
func (m *mockGroupRepo) MuteMember(ctx context.Context, groupID, userID string, seconds int32) error {
	m.mutes[groupID+":"+userID] = true
	return nil
}
func (m *mockGroupRepo) UnmuteMember(ctx context.Context, groupID, userID string) error {
	delete(m.mutes, groupID+":"+userID)
	return nil
}
func (m *mockGroupRepo) IsMuted(ctx context.Context, groupID, userID string) (bool, error) {
	return m.mutes[groupID+":"+userID], nil
}

func TestCreateGroup(t *testing.T) {
	repo := newMockGroupRepo()
	uc := NewGroupUseCase(repo)

	group, err := uc.CreateGroup(context.Background(), "Test Group", "owner-1")
	if err != nil {
		t.Fatalf("CreateGroup: %v", err)
	}
	if group.GroupID == "" {
		t.Fatal("expected group ID")
	}
}

func TestJoinAndLeaveGroup(t *testing.T) {
	repo := newMockGroupRepo()
	uc := NewGroupUseCase(repo)

	group, _ := uc.CreateGroup(context.Background(), "G1", "owner")

	err := uc.JoinGroup(context.Background(), group.GroupID, "member-1")
	if err != nil {
		t.Fatalf("JoinGroup: %v", err)
	}

	members, _ := uc.GetGroupMembers(context.Background(), group.GroupID, 0, 50)
	if len(members) != 2 {
		t.Fatalf("expected 2 members, got %d", len(members))
	}

	err = uc.LeaveGroup(context.Background(), group.GroupID, "member-1")
	if err != nil {
		t.Fatalf("LeaveGroup: %v", err)
	}
}

func TestDismissGroup(t *testing.T) {
	repo := newMockGroupRepo()
	uc := NewGroupUseCase(repo)

	group, _ := uc.CreateGroup(context.Background(), "ToDelete", "owner")

	err := uc.DismissGroup(context.Background(), group.GroupID, "owner")
	if err != nil {
		t.Fatalf("DismissGroup: %v", err)
	}
}

func TestDismissGroupNotOwner(t *testing.T) {
	repo := newMockGroupRepo()
	uc := NewGroupUseCase(repo)

	group, _ := uc.CreateGroup(context.Background(), "ToDelete", "owner")

	err := uc.DismissGroup(context.Background(), group.GroupID, "not-owner")
	if err == nil {
		t.Fatal("expected error for non-owner dismiss")
	}
}

func TestMuteAndUnmute(t *testing.T) {
	repo := newMockGroupRepo()
	uc := NewGroupUseCase(repo)

	group, _ := uc.CreateGroup(context.Background(), "G1", "owner")
	uc.JoinGroup(context.Background(), group.GroupID, "member")

	err := uc.MuteMember(context.Background(), group.GroupID, "member", "owner", 3600)
	if err != nil {
		t.Fatalf("MuteMember: %v", err)
	}

	muted, _ := uc.IsMuted(context.Background(), group.GroupID, "member")
	if !muted {
		t.Fatal("expected member to be muted")
	}

	err = uc.UnmuteMember(context.Background(), group.GroupID, "member")
	if err != nil {
		t.Fatalf("UnmuteMember: %v", err)
	}
}

func TestMuteMemberNotOwner(t *testing.T) {
	repo := newMockGroupRepo()
	uc := NewGroupUseCase(repo)

	group, _ := uc.CreateGroup(context.Background(), "G1", "owner")

	err := uc.MuteMember(context.Background(), group.GroupID, "member", "not-owner", 3600)
	if err == nil {
		t.Fatal("expected error for non-owner mute")
	}
}

func TestKickMember(t *testing.T) {
	repo := newMockGroupRepo()
	uc := NewGroupUseCase(repo)

	group, _ := uc.CreateGroup(context.Background(), "G1", "owner")
	uc.JoinGroup(context.Background(), group.GroupID, "bad-member")

	err := uc.KickMember(context.Background(), group.GroupID, "bad-member", "owner")
	if err != nil {
		t.Fatalf("KickMember: %v", err)
	}
}
