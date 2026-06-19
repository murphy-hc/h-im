package biz

import (
	"sync"

	"github.com/gorilla/websocket"
)

type ConnManager interface {
	Add(userID string, conn *websocket.Conn)
	Remove(userID string)
	GetConn(userID string) (*websocket.Conn, bool)
	GetGroupMembers(groupID string) []string
	JoinGroup(groupID, userID string)
	LeaveGroup(groupID, userID string)
	GetRoomMembers(roomID string) []string
	JoinRoom(roomID, userID string)
	LeaveRoom(roomID, userID string)
	OnlineCount() int
}

type connManager struct {
	mu         sync.RWMutex
	conns      map[string]*websocket.Conn
	groupUsers map[string]map[string]struct{}
	roomUsers  map[string]map[string]struct{}
	userGroups map[string]map[string]struct{}
	userRooms  map[string]map[string]struct{}
}

func NewConnManager() ConnManager {
	return &connManager{
		conns:      make(map[string]*websocket.Conn),
		groupUsers: make(map[string]map[string]struct{}),
		roomUsers:  make(map[string]map[string]struct{}),
		userGroups: make(map[string]map[string]struct{}),
		userRooms:  make(map[string]map[string]struct{}),
	}
}

func (cm *connManager) Add(userID string, conn *websocket.Conn) {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	cm.conns[userID] = conn
}

func (cm *connManager) Remove(userID string) {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	delete(cm.conns, userID)
	for gid := range cm.userGroups[userID] {
		delete(cm.groupUsers[gid], userID)
	}
	for rid := range cm.userRooms[userID] {
		delete(cm.roomUsers[rid], userID)
	}
	delete(cm.userGroups, userID)
	delete(cm.userRooms, userID)
}

func (cm *connManager) GetConn(userID string) (*websocket.Conn, bool) {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	c, ok := cm.conns[userID]
	return c, ok
}

func (cm *connManager) JoinGroup(groupID, userID string) {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	if cm.groupUsers[groupID] == nil {
		cm.groupUsers[groupID] = make(map[string]struct{})
	}
	cm.groupUsers[groupID][userID] = struct{}{}
	if cm.userGroups[userID] == nil {
		cm.userGroups[userID] = make(map[string]struct{})
	}
	cm.userGroups[userID][groupID] = struct{}{}
}

func (cm *connManager) LeaveGroup(groupID, userID string) {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	delete(cm.groupUsers[groupID], userID)
	delete(cm.userGroups[userID], groupID)
}

func (cm *connManager) GetGroupMembers(groupID string) []string {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	var ids []string
	for id := range cm.groupUsers[groupID] {
		ids = append(ids, id)
	}
	return ids
}

func (cm *connManager) JoinRoom(roomID, userID string) {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	if cm.roomUsers[roomID] == nil {
		cm.roomUsers[roomID] = make(map[string]struct{})
	}
	cm.roomUsers[roomID][userID] = struct{}{}
	if cm.userRooms[userID] == nil {
		cm.userRooms[userID] = make(map[string]struct{})
	}
	cm.userRooms[userID][roomID] = struct{}{}
}

func (cm *connManager) LeaveRoom(roomID, userID string) {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	delete(cm.roomUsers[roomID], userID)
	delete(cm.userRooms[userID], roomID)
}

func (cm *connManager) GetRoomMembers(roomID string) []string {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	var ids []string
	for id := range cm.roomUsers[roomID] {
		ids = append(ids, id)
	}
	return ids
}

func (cm *connManager) OnlineCount() int {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	return len(cm.conns)
}
