package biz

import "context"

// UserStatusClient reports heartbeat and disconnect events to the user service.
type UserStatusClient interface {
	ReportHeartbeat(ctx context.Context, userID, deviceID, gatewayAddr string, timestamp int64) error
	ReportDisconnect(ctx context.Context, userID, deviceID string) error
}
