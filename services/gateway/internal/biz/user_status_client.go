package biz

import "context"

// UserStatusClient communicates with the user service.
type UserStatusClient interface {
	ReportHeartbeat(ctx context.Context, userID, deviceID, gatewayAddr string, timestamp int64) error
	ReportDisconnect(ctx context.Context, userID, deviceID string) error
	ValidateAppToken(ctx context.Context, appID, userID, token string) (bool, error)
}
