package biz

import "context"

// GroupClient is the interface for calling the group service.
type GroupClient interface {
	JoinGroup(ctx context.Context, groupID, userID string) error
	LeaveGroup(ctx context.Context, groupID, userID string) error
}
