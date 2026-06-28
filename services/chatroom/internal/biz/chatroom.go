package biz

import (
	"context"
	"fmt"

	"github.com/rs/xid"
)

// ChatroomUseCase handles chatroom business logic.
type ChatroomUseCase struct {
	repo ChatroomRepo
}

// NewChatroomUseCase creates a ChatroomUseCase.
func NewChatroomUseCase(repo ChatroomRepo) *ChatroomUseCase {
	return &ChatroomUseCase{repo: repo}
}

func (uc *ChatroomUseCase) CreateRoom(ctx context.Context, name, ownerID string) (*Room, error) {
	roomID := xid.New().String()
	if err := uc.repo.CreateRoom(ctx, roomID, name, ownerID); err != nil {
		return nil, fmt.Errorf("create room: %w", err)
	}
	return &Room{RoomID: roomID, Name: name, OwnerID: ownerID, MemberCount: 1}, nil
}

func (uc *ChatroomUseCase) JoinRoom(ctx context.Context, roomID, userID string) error {
	if _, err := uc.repo.FindByID(ctx, roomID); err != nil {
		return fmt.Errorf("room not found: %w", err)
	}
	return uc.repo.JoinRoom(ctx, roomID, userID)
}

func (uc *ChatroomUseCase) LeaveRoom(ctx context.Context, roomID, userID string) error {
	return uc.repo.LeaveRoom(ctx, roomID, userID)
}

func (uc *ChatroomUseCase) GetMembers(ctx context.Context, roomID string) ([]string, error) {
	return uc.repo.GetMembers(ctx, roomID)
}

func (uc *ChatroomUseCase) GetMessages(ctx context.Context, roomID string, offset, limit int32) ([]*ChatroomMessage, int64, error) {
	return uc.repo.GetMessages(ctx, roomID, offset, limit)
}
