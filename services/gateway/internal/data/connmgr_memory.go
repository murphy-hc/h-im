package data

import (
	"sync"

	"github.com/gorilla/websocket"
	"github.com/murphy-hc/h-im/services/gateway/internal/biz"
)

// memoryConnManager is an in-memory ConnManager for testing.
type memoryConnManager struct {
	mu         sync.RWMutex
	localConns map[string]map[string]*websocket.Conn // userID -> deviceID -> conn
	groupUsers map[string]map[string]struct{}
	roomUsers  map[string]map[string]struct{}
}

func newMemoryConnManager() biz.ConnManager {
	return &memoryConnManager{
		localConns: make(map[string]map[string]*websocket.Conn),
		groupUsers: make(map[string]map[string]struct{}),
		roomUsers:  make(map[string]map[string]struct{}),
	}
}

func (cm *memoryConnManager) Add(userID, deviceID string, conn *websocket.Conn) {
	cm.mu.Lock(); defer cm.mu.Unlock()
	if cm.localConns[userID] == nil { cm.localConns[userID] = make(map[string]*websocket.Conn) }
	cm.localConns[userID][deviceID] = conn
}
func (cm *memoryConnManager) Remove(userID, deviceID string) {
	cm.mu.Lock(); defer cm.mu.Unlock()
	if devs, ok := cm.localConns[userID]; ok {
		delete(devs, deviceID)
		if len(devs) == 0 { delete(cm.localConns, userID) }
	}
}
func (cm *memoryConnManager) GetConns(userID string) []*websocket.Conn {
	cm.mu.RLock(); defer cm.mu.RUnlock()
	devs := cm.localConns[userID]
	if len(devs) == 0 { return nil }
	conns := make([]*websocket.Conn, 0, len(devs))
	for _, c := range devs { conns = append(conns, c) }
	return conns
}
func (cm *memoryConnManager) KickUser(userID string) []*websocket.Conn {
	cm.mu.Lock(); defer cm.mu.Unlock()
	devs := cm.localConns[userID]
	delete(cm.localConns, userID)
	conns := make([]*websocket.Conn, 0, len(devs))
	for _, c := range devs { conns = append(conns, c) }
	return conns
}
func (cm *memoryConnManager) JoinGroup(groupID, userID string) {
	cm.mu.Lock(); defer cm.mu.Unlock()
	if cm.groupUsers[groupID] == nil { cm.groupUsers[groupID] = make(map[string]struct{}) }
	cm.groupUsers[groupID][userID] = struct{}{}
}
func (cm *memoryConnManager) LeaveGroup(groupID, userID string) {
	cm.mu.Lock(); defer cm.mu.Unlock()
	delete(cm.groupUsers[groupID], userID)
}
func (cm *memoryConnManager) GetGroupMembers(groupID string) []string {
	cm.mu.RLock(); defer cm.mu.RUnlock()
	var ids []string
	for id := range cm.groupUsers[groupID] { ids = append(ids, id) }
	return ids
}
func (cm *memoryConnManager) JoinRoom(roomID, userID string) {
	cm.mu.Lock(); defer cm.mu.Unlock()
	if cm.roomUsers[roomID] == nil { cm.roomUsers[roomID] = make(map[string]struct{}) }
	cm.roomUsers[roomID][userID] = struct{}{}
}
func (cm *memoryConnManager) LeaveRoom(roomID, userID string) {
	cm.mu.Lock(); defer cm.mu.Unlock()
	delete(cm.roomUsers[roomID], userID)
}
func (cm *memoryConnManager) GetRoomMembers(roomID string) []string {
	cm.mu.RLock(); defer cm.mu.RUnlock()
	var ids []string
	for id := range cm.roomUsers[roomID] { ids = append(ids, id) }
	return ids
}
func (cm *memoryConnManager) OnlineCount() int {
	cm.mu.RLock(); defer cm.mu.RUnlock()
	total := 0
	for _, devs := range cm.localConns { total += len(devs) }
	return total
}
