package biz

import (
	"context"
	"testing"
)

type mockChatroomRepo struct {
	rooms    map[string]*Room
	members  map[string][]string
	messages []*ChatroomMessage
}

func newMockChatroomRepo() *mockChatroomRepo {
	return &mockChatroomRepo{
		rooms:   make(map[string]*Room),
		members: make(map[string][]string),
	}
}

func (m *mockChatroomRepo) CreateRoom(ctx context.Context, roomID, name, ownerID string) error {
	m.rooms[roomID] = &Room{RoomID: roomID, Name: name, OwnerID: ownerID, MemberCount: 1}
	m.members[roomID] = []string{ownerID}
	return nil
}
func (m *mockChatroomRepo) FindByID(ctx context.Context, roomID string) (*Room, error) {
	r, ok := m.rooms[roomID]
	if !ok {
		return nil, context.DeadlineExceeded
	}
	return r, nil
}
func (m *mockChatroomRepo) JoinRoom(ctx context.Context, roomID, userID string) error {
	m.members[roomID] = append(m.members[roomID], userID)
	if r, ok := m.rooms[roomID]; ok {
		r.MemberCount++
	}
	return nil
}
func (m *mockChatroomRepo) LeaveRoom(ctx context.Context, roomID, userID string) error {
	return nil
}
func (m *mockChatroomRepo) GetMembers(ctx context.Context, roomID string) ([]string, error) {
	return m.members[roomID], nil
}
func (m *mockChatroomRepo) GetMessages(ctx context.Context, roomID string, offset, limit int32) ([]*ChatroomMessage, int64, error) {
	return m.messages, int64(len(m.messages)), nil
}

func TestCreateRoom(t *testing.T) {
	repo := newMockChatroomRepo()
	uc := NewChatroomUseCase(repo)

	room, err := uc.CreateRoom(context.Background(), "General", "owner-1")
	if err != nil {
		t.Fatalf("CreateRoom: %v", err)
	}
	if room.RoomID == "" {
		t.Fatal("expected room ID")
	}
	if room.Name != "General" {
		t.Fatalf("expected name 'General', got %q", room.Name)
	}
}

func TestJoinRoom(t *testing.T) {
	repo := newMockChatroomRepo()
	uc := NewChatroomUseCase(repo)

	room, _ := uc.CreateRoom(context.Background(), "Lobby", "owner")
	err := uc.JoinRoom(context.Background(), room.RoomID, "user-2")
	if err != nil {
		t.Fatalf("JoinRoom: %v", err)
	}

	members, _ := uc.GetMembers(context.Background(), room.RoomID)
	if len(members) != 2 {
		t.Fatalf("expected 2 members, got %d", len(members))
	}
}

func TestJoinRoomNotFound(t *testing.T) {
	repo := newMockChatroomRepo()
	uc := NewChatroomUseCase(repo)

	err := uc.JoinRoom(context.Background(), "nonexistent", "user")
	if err == nil {
		t.Fatal("expected error for non-existent room")
	}
}

func TestLeaveRoom(t *testing.T) {
	repo := newMockChatroomRepo()
	uc := NewChatroomUseCase(repo)

	room, _ := uc.CreateRoom(context.Background(), "Lobby", "owner")
	uc.JoinRoom(context.Background(), room.RoomID, "user-2")

	err := uc.LeaveRoom(context.Background(), room.RoomID, "user-2")
	if err != nil {
		t.Fatalf("LeaveRoom: %v", err)
	}
}

func TestGetMessages(t *testing.T) {
	repo := newMockChatroomRepo()
	uc := NewChatroomUseCase(repo)

	room, _ := uc.CreateRoom(context.Background(), "Lobby", "owner")
	repo.messages = append(repo.messages, &ChatroomMessage{
		ServerID: "1", RoomID: room.RoomID, SenderID: "user-1", Text: "hello",
	})

	msgs, total, err := uc.GetMessages(context.Background(), room.RoomID, 0, 20)
	if err != nil {
		t.Fatalf("GetMessages: %v", err)
	}
	if total != 1 {
		t.Fatalf("expected 1 message, got %d", total)
	}
	if len(msgs) != 1 {
		t.Fatalf("expected 1 message, got %d", len(msgs))
	}
}
