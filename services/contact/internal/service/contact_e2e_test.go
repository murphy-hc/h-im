package service

import (
	"context"
	"testing"

	pb "github.com/murphy-hc/h-im/gen/go/him/contact/v1"
	"github.com/murphy-hc/h-im/services/contact/internal/biz"
)

type e2eContactRepo struct {
	requests map[string]*biz.FriendRequest
	friends  map[string]map[string]int32
}

func newE2EContactRepo() *e2eContactRepo {
	return &e2eContactRepo{
		requests: make(map[string]*biz.FriendRequest),
		friends:  make(map[string]map[string]int32),
	}
}

func (r *e2eContactRepo) SendRequest(_ context.Context, reqID, from, to, msg string) error {
	r.requests[reqID] = &biz.FriendRequest{RequestID: reqID, FromUser: from, ToUser: to, Message: msg, Status: 0}
	return nil
}
func (r *e2eContactRepo) RespondRequest(_ context.Context, reqID string, accept bool) error {
	req := r.requests[reqID]
	if req == nil {
		return context.DeadlineExceeded
	}
	if accept {
		req.Status = 1
		if r.friends[req.FromUser] == nil {
			r.friends[req.FromUser] = make(map[string]int32)
		}
		if r.friends[req.ToUser] == nil {
			r.friends[req.ToUser] = make(map[string]int32)
		}
		r.friends[req.FromUser][req.ToUser] = 1
		r.friends[req.ToUser][req.FromUser] = 1
	} else {
		req.Status = 2
	}
	return nil
}
func (r *e2eContactRepo) FindPendingRequest(_ context.Context, _, _ string) (*biz.FriendRequest, error) {
	return nil, context.DeadlineExceeded
}
func (r *e2eContactRepo) RemoveFriend(_ context.Context, userID, friendID string) error {
	delete(r.friends[userID], friendID)
	delete(r.friends[friendID], userID)
	return nil
}
func (r *e2eContactRepo) BlockUser(_ context.Context, userID, blockID string) error {
	if r.friends[userID] == nil {
		r.friends[userID] = make(map[string]int32)
	}
	r.friends[userID][blockID] = 2
	return nil
}
func (r *e2eContactRepo) UnblockUser(_ context.Context, userID, blockID string) error {
	if r.friends[userID] == nil {
		r.friends[userID] = make(map[string]int32)
	}
	r.friends[userID][blockID] = 1
	return nil
}
func (r *e2eContactRepo) GetFriends(_ context.Context, userID string, _, _ int32) ([]biz.FriendInfo, error) {
	var out []biz.FriendInfo
	for id, status := range r.friends[userID] {
		out = append(out, biz.FriendInfo{UserID: id, Status: status})
	}
	return out, nil
}
func (r *e2eContactRepo) GetRequests(_ context.Context, userID string) ([]biz.FriendRequest, error) {
	var out []biz.FriendRequest
	for _, req := range r.requests {
		if req.ToUser == userID && req.Status == 0 {
			out = append(out, *req)
		}
	}
	return out, nil
}

func newE2EContactService() *ContactService {
	return NewContactService(biz.NewContactUseCase(newE2EContactRepo()))
}

func TestE2E_FullFriendFlow(t *testing.T) {
	svc := newE2EContactService()
	ctx := context.Background()

	// 1. Send friend request
	reqResp, err := svc.SendFriendRequest(ctx, &pb.SendFriendRequestRequest{
		FromUser: "alice", ToUser: "bob", Message: "Hi Bob",
	})
	if err != nil {
		t.Fatalf("SendFriendRequest: %v", err)
	}
	if reqResp.RequestId == "" {
		t.Fatal("expected request ID")
	}

	// 2. Accept
	_, err = svc.AcceptFriendRequest(ctx, &pb.AcceptFriendRequestRequest{RequestId: reqResp.RequestId})
	if err != nil {
		t.Fatalf("AcceptFriendRequest: %v", err)
	}

	// 3. Check friends
	friends, err := svc.GetFriends(ctx, &pb.GetFriendsRequest{UserId: "alice"})
	if err != nil {
		t.Fatalf("GetFriends: %v", err)
	}
	if len(friends.Friends) != 1 || friends.Friends[0].UserId != "bob" {
		t.Fatal("expected bob as friend")
	}

	// 4. Remove friend
	_, err = svc.RemoveFriend(ctx, &pb.RemoveFriendRequest{UserId: "alice", FriendId: "bob"})
	if err != nil {
		t.Fatalf("RemoveFriend: %v", err)
	}
}

func TestE2E_SelfFriendRequest(t *testing.T) {
	svc := newE2EContactService()
	ctx := context.Background()

	_, err := svc.SendFriendRequest(ctx, &pb.SendFriendRequestRequest{
		FromUser: "alice", ToUser: "alice", Message: "me",
	})
	if err == nil {
		t.Fatal("expected error for self-friend")
	}
}
