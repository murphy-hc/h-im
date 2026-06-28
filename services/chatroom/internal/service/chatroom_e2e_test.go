package service

import (
	"context"
	"testing"

	pb "github.com/murphy-hc/h-im/gen/go/him/chatroom/v1"
	commonv1 "github.com/murphy-hc/h-im/gen/go/him/common/v1"
	"github.com/murphy-hc/h-im/services/chatroom/internal/biz"
)

type e2eChatroomRepo struct {
	rooms    map[string]*biz.Room
	members  map[string][]string
	messages []*biz.ChatroomMessage
}

func newE2EChatroomRepo() *e2eChatroomRepo {
	return &e2eChatroomRepo{
		rooms:   make(map[string]*biz.Room),
		members: make(map[string][]string),
	}
}

func (r *e2eChatroomRepo) CreateRoom(_ context.Context, roomID, name, ownerID string) error {
	r.rooms[roomID] = &biz.Room{RoomID: roomID, Name: name, OwnerID: ownerID, MemberCount: 1}
	r.members[roomID] = []string{ownerID}
	return nil
}
func (r *e2eChatroomRepo) FindByID(_ context.Context, roomID string) (*biz.Room, error) {
	room, ok := r.rooms[roomID]
	if !ok {
		return nil, context.DeadlineExceeded
	}
	return room, nil
}
func (r *e2eChatroomRepo) JoinRoom(_ context.Context, roomID, userID string) error {
	r.members[roomID] = append(r.members[roomID], userID)
	if room, ok := r.rooms[roomID]; ok {
		room.MemberCount++
	}
	return nil
}
func (r *e2eChatroomRepo) LeaveRoom(_ context.Context, _, _ string) error { return nil }
func (r *e2eChatroomRepo) GetMembers(_ context.Context, roomID string) ([]string, error) {
	return r.members[roomID], nil
}
func (r *e2eChatroomRepo) GetMessages(_ context.Context, _ string, _, _ int32) ([]*biz.ChatroomMessage, int64, error) {
	return r.messages, int64(len(r.messages)), nil
}

func newE2EChatroomService() *ChatroomService {
	return NewChatroomService(biz.NewChatroomUseCase(newE2EChatroomRepo()))
}

func TestE2E_CreateJoinLeave(t *testing.T) {
	svc := newE2EChatroomService()
	ctx := context.Background()

	// 1. Create
	createResp, err := svc.CreateRoom(ctx, &pb.CreateRoomRequest{Name: "Lobby", OwnerId: "user-1"})
	if err != nil {
		t.Fatalf("CreateRoom: %v", err)
	}
	rid := createResp.RoomId

	// 2. Join
	_, err = svc.JoinRoom(ctx, &pb.JoinRoomRequest{RoomId: rid, UserId: "user-2"})
	if err != nil {
		t.Fatalf("JoinRoom: %v", err)
	}

	// 3. Leave
	_, err = svc.LeaveRoom(ctx, &pb.LeaveRoomRequest{RoomId: rid, UserId: "user-2"})
	if err != nil {
		t.Fatalf("LeaveRoom: %v", err)
	}
}

func TestE2E_GetMessages(t *testing.T) {
	svc := newE2EChatroomService()
	ctx := context.Background()

	createResp, _ := svc.CreateRoom(ctx, &pb.CreateRoomRequest{Name: "TestRoom", OwnerId: "user-1"})

	resp, err := svc.GetMessages(ctx, &pb.GetMessagesRequest{
		RoomId: createResp.RoomId,
		Pagination: &commonv1.Pagination{Page: 1, PageSize: 20},
	})
	if err != nil {
		t.Fatalf("GetMessages: %v", err)
	}
	if resp.Pagination == nil {
		t.Fatal("expected pagination in response")
	}
}
