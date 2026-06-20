package biz

import "github.com/gorilla/websocket"

type ConnManager interface {
	Add(userID, deviceID string, conn *websocket.Conn)
	Remove(userID, deviceID string)
	GetConns(userID string) []*websocket.Conn
	KickUser(userID string) []*websocket.Conn
	GetGroupMembers(groupID string) []string
	JoinGroup(groupID, userID string)
	LeaveGroup(groupID, userID string)
	GetRoomMembers(roomID string) []string
	JoinRoom(roomID, userID string)
	LeaveRoom(roomID, userID string)
	OnlineCount() int
}
