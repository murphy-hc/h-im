package biz

import (
	"context"
	"time"

	"github.com/coder/websocket"
)

// WebSocket close reasons.
const (
	CloseReasonAuthFailed       = "auth failed"
	CloseReasonKicked           = "kicked"
	CloseReasonHeartbeatTimeout = "heartbeat timeout"
)

// ConnState tracks per-connection heartbeat state.
type ConnState struct {
	Conn                    *websocket.Conn
	LastSuccessHeartbeat    time.Time
	ConsecutiveEchoFailures int
}

// OfflineDevice identifies a connection that has timed out.
type OfflineDevice struct {
	UserID   string
	DeviceID string
	Conn     *websocket.Conn
}

type ConnManager interface {
	Add(ctx context.Context, userID, deviceID string, conn *websocket.Conn) error
	Remove(ctx context.Context, userID, deviceID string) error
	GetConns(ctx context.Context, userID string) ([]*websocket.Conn, error)
	KickUser(ctx context.Context, userID string) ([]*websocket.Conn, error)
	GetGroupMembers(ctx context.Context, groupID string) ([]string, error)
	JoinGroup(ctx context.Context, groupID, userID string) error
	LeaveGroup(ctx context.Context, groupID, userID string) error
	GetRoomMembers(ctx context.Context, roomID string) ([]string, error)
	JoinRoom(ctx context.Context, roomID, userID string) error
	LeaveRoom(ctx context.Context, roomID, userID string) error
	OnlineCount() int

	// MarkHeartbeatSuccess marks a successful heartbeat (echo sent).
	MarkHeartbeatSuccess(userID, deviceID string)
	// MarkHeartbeatFail records a failed heartbeat echo.
	MarkHeartbeatFail(userID, deviceID string)
	// SweepOffline scans all connections and returns those that have exceeded the timeout.
	// It also removes the connections from the local map.
	SweepOffline(ctx context.Context, timeout time.Duration) []OfflineDevice
}
