package biz

import "github.com/gorilla/websocket"

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
}
