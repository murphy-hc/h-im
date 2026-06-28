package biz

import (
	"context"
	"fmt"

	"github.com/rs/xid"
)

// ContactUseCase handles contact business logic.
type ContactUseCase struct {
	repo ContactRepo
}

func NewContactUseCase(repo ContactRepo) *ContactUseCase {
	return &ContactUseCase{repo: repo}
}

func (uc *ContactUseCase) SendFriendRequest(ctx context.Context, fromUser, toUser, msg string) (string, error) {
	if fromUser == toUser {
		return "", fmt.Errorf("cannot friend yourself")
	}
	req, _ := uc.repo.FindPendingRequest(ctx, fromUser, toUser)
	if req != nil {
		return "", fmt.Errorf("request already sent")
	}
	reqID := xid.New().String()
	if err := uc.repo.SendRequest(ctx, reqID, fromUser, toUser, msg); err != nil {
		return "", fmt.Errorf("send: %w", err)
	}
	return reqID, nil
}

func (uc *ContactUseCase) AcceptFriendRequest(ctx context.Context, reqID string) error {
	return uc.repo.RespondRequest(ctx, reqID, true)
}

func (uc *ContactUseCase) RejectFriendRequest(ctx context.Context, reqID string) error {
	return uc.repo.RespondRequest(ctx, reqID, false)
}

func (uc *ContactUseCase) RemoveFriend(ctx context.Context, userID, friendID string) error {
	return uc.repo.RemoveFriend(ctx, userID, friendID)
}

func (uc *ContactUseCase) BlockUser(ctx context.Context, userID, blockID string) error {
	return uc.repo.BlockUser(ctx, userID, blockID)
}

func (uc *ContactUseCase) UnblockUser(ctx context.Context, userID, blockID string) error {
	return uc.repo.UnblockUser(ctx, userID, blockID)
}

func (uc *ContactUseCase) GetFriends(ctx context.Context, userID string, offset, limit int32) ([]FriendInfo, error) {
	return uc.repo.GetFriends(ctx, userID, offset, limit)
}

func (uc *ContactUseCase) GetFriendRequests(ctx context.Context, userID string) ([]FriendRequest, error) {
	return uc.repo.GetRequests(ctx, userID)
}
