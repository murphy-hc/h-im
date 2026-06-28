package biz

import (
	"context"
	"testing"
)

type mockContactRepo struct {
	friends    map[string]map[string]int32
	requests   map[string]*FriendRequest
}

func newMockContactRepo() *mockContactRepo {
	return &mockContactRepo{
		friends:  make(map[string]map[string]int32),
		requests: make(map[string]*FriendRequest),
	}
}

func (m *mockContactRepo) SendRequest(ctx context.Context, reqID, fromUser, toUser, msg string) error {
	m.requests[reqID] = &FriendRequest{RequestID: reqID, FromUser: fromUser, ToUser: toUser, Message: msg, Status: 0}
	return nil
}
func (m *mockContactRepo) RespondRequest(ctx context.Context, reqID string, accept bool) error {
	r, ok := m.requests[reqID]
	if !ok {
		return context.DeadlineExceeded
	}
	if accept {
		r.Status = 1
		m.friends[r.FromUser] = map[string]int32{r.ToUser: 1}
		m.friends[r.ToUser] = map[string]int32{r.FromUser: 1}
	} else {
		r.Status = 2
	}
	return nil
}
func (m *mockContactRepo) FindPendingRequest(ctx context.Context, fromUser, toUser string) (*FriendRequest, error) {
	return nil, context.DeadlineExceeded
}
func (m *mockContactRepo) RemoveFriend(ctx context.Context, userID, friendID string) error {
	delete(m.friends[userID], friendID)
	delete(m.friends[friendID], userID)
	return nil
}
func (m *mockContactRepo) BlockUser(ctx context.Context, userID, blockID string) error {
	if m.friends[userID] == nil {
		m.friends[userID] = make(map[string]int32)
	}
	m.friends[userID][blockID] = 2
	return nil
}
func (m *mockContactRepo) UnblockUser(ctx context.Context, userID, blockID string) error {
	if m.friends[userID] == nil {
		m.friends[userID] = make(map[string]int32)
	}
	m.friends[userID][blockID] = 1
	return nil
}
func (m *mockContactRepo) GetFriends(ctx context.Context, userID string, offset, limit int32) ([]FriendInfo, error) {
	var out []FriendInfo
	fMap := m.friends[userID]
	for id, status := range fMap {
		out = append(out, FriendInfo{UserID: id, Status: status})
	}
	return out, nil
}
func (m *mockContactRepo) GetRequests(ctx context.Context, userID string) ([]FriendRequest, error) {
	var out []FriendRequest
	for _, r := range m.requests {
		if r.ToUser == userID && r.Status == 0 {
			out = append(out, *r)
		}
	}
	return out, nil
}

func TestSendFriendRequest(t *testing.T) {
	repo := newMockContactRepo()
	uc := NewContactUseCase(repo)

	id, err := uc.SendFriendRequest(context.Background(), "alice", "bob", "hello")
	if err != nil {
		t.Fatalf("SendFriendRequest: %v", err)
	}
	if id == "" {
		t.Fatal("expected request ID")
	}
}

func TestSendFriendRequestToSelf(t *testing.T) {
	repo := newMockContactRepo()
	uc := NewContactUseCase(repo)

	_, err := uc.SendFriendRequest(context.Background(), "alice", "alice", "hi")
	if err == nil {
		t.Fatal("expected error for self-friend request")
	}
}

func TestAcceptFriendRequest(t *testing.T) {
	repo := newMockContactRepo()
	uc := NewContactUseCase(repo)

	id, _ := uc.SendFriendRequest(context.Background(), "alice", "bob", "hello")
	err := uc.AcceptFriendRequest(context.Background(), id)
	if err != nil {
		t.Fatalf("AcceptFriendRequest: %v", err)
	}

	friends, _ := uc.GetFriends(context.Background(), "alice", 0, 50)
	if len(friends) != 1 || friends[0].UserID != "bob" {
		t.Fatalf("expected bob as friend, got %v", friends)
	}
}

func TestRejectFriendRequest(t *testing.T) {
	repo := newMockContactRepo()
	uc := NewContactUseCase(repo)

	id, _ := uc.SendFriendRequest(context.Background(), "alice", "bob", "hello")
	err := uc.RejectFriendRequest(context.Background(), id)
	if err != nil {
		t.Fatalf("RejectFriendRequest: %v", err)
	}
}

func TestRemoveFriend(t *testing.T) {
	repo := newMockContactRepo()
	uc := NewContactUseCase(repo)

	id, _ := uc.SendFriendRequest(context.Background(), "alice", "bob", "hello")
	uc.AcceptFriendRequest(context.Background(), id)

	err := uc.RemoveFriend(context.Background(), "alice", "bob")
	if err != nil {
		t.Fatalf("RemoveFriend: %v", err)
	}
}

func TestBlockAndUnblock(t *testing.T) {
	repo := newMockContactRepo()
	uc := NewContactUseCase(repo)

	err := uc.BlockUser(context.Background(), "alice", "spammer")
	if err != nil {
		t.Fatalf("BlockUser: %v", err)
	}

	err = uc.UnblockUser(context.Background(), "alice", "spammer")
	if err != nil {
		t.Fatalf("UnblockUser: %v", err)
	}
}
