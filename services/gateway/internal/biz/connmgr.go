package biz

import (
	"time"

	"github.com/coder/websocket"
)

// WebSocket close reasons.
const (
	CloseReasonAuthFailed      = "auth failed"
	CloseReasonKicked          = "kicked"
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
	Add(userID, deviceID string, conn *websocket.Conn) error
	Remove(userID, deviceID string) error
	GetConns(userID string) ([]*websocket.Conn, error)
	KickUser(userID string) ([]*websocket.Conn, error)
	GetGroupMembers(groupID string) ([]string, error)
	JoinGroup(groupID, userID string) error
	LeaveGroup(groupID, userID string) error
	GetRoomMembers(roomID string) ([]string, error)
	JoinRoom(roomID, userID string) error
	LeaveRoom(roomID, userID string) error
	OnlineCount() int

	// MarkHeartbeatSuccess marks a successful heartbeat (echo sent).
	MarkHeartbeatSuccess(userID, deviceID string)
	// MarkHeartbeatFail records a failed heartbeat echo.
	MarkHeartbeatFail(userID, deviceID string)
	// SweepOffline scans all connections and returns those that have exceeded the timeout.
	// It also removes the connections from the local map.
	SweepOffline(timeout time.Duration) []OfflineDevice
}
