package data

import (
	"context"
	"sync"
	"time"

	"github.com/coder/websocket"
	"github.com/murphy-hc/h-im/services/gateway/internal/biz"
)

type memoryConnManager struct {
	mu         sync.RWMutex
	localConns map[string]map[string]*biz.ConnState
	groupUsers map[string]map[string]struct{}
	roomUsers  map[string]map[string]struct{}
}

func newMemoryConnManager() biz.ConnManager {
	return &memoryConnManager{
		localConns: make(map[string]map[string]*biz.ConnState),
		groupUsers: make(map[string]map[string]struct{}),
		roomUsers:  make(map[string]map[string]struct{}),
	}
}

func (cm *memoryConnManager) Add(_ context.Context, userID, deviceID string, conn *websocket.Conn) error {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	if cm.localConns[userID] == nil {
		cm.localConns[userID] = make(map[string]*biz.ConnState)
	}
	cm.localConns[userID][deviceID] = &biz.ConnState{
		Conn:                 conn,
		LastSuccessHeartbeat: time.Now(),
	}
	return nil
}

func (cm *memoryConnManager) Remove(_ context.Context, userID, deviceID string) error {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	if devs, ok := cm.localConns[userID]; ok {
		delete(devs, deviceID)
		if len(devs) == 0 {
			delete(cm.localConns, userID)
		}
	}
	return nil
}

func (cm *memoryConnManager) GetConns(_ context.Context, userID string) ([]*websocket.Conn, error) {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	devs := cm.localConns[userID]
	if len(devs) == 0 {
		return nil, nil
	}
	conns := make([]*websocket.Conn, 0, len(devs))
	for _, cs := range devs {
		conns = append(conns, cs.Conn)
	}
	return conns, nil
}

func (cm *memoryConnManager) KickUser(_ context.Context, userID string) ([]*websocket.Conn, error) {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	devs := cm.localConns[userID]
	delete(cm.localConns, userID)
	conns := make([]*websocket.Conn, 0, len(devs))
	for _, cs := range devs {
		conns = append(conns, cs.Conn)
	}
	return conns, nil
}

func (cm *memoryConnManager) JoinGroup(_ context.Context, groupID, userID string) error {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	if cm.groupUsers[groupID] == nil {
		cm.groupUsers[groupID] = make(map[string]struct{})
	}
	cm.groupUsers[groupID][userID] = struct{}{}
	return nil
}

func (cm *memoryConnManager) LeaveGroup(_ context.Context, groupID, userID string) error {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	delete(cm.groupUsers[groupID], userID)
	return nil
}

func (cm *memoryConnManager) GetGroupMembers(_ context.Context, groupID string) ([]string, error) {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	var ids []string
	for id := range cm.groupUsers[groupID] {
		ids = append(ids, id)
	}
	return ids, nil
}

func (cm *memoryConnManager) JoinRoom(_ context.Context, roomID, userID string) error {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	if cm.roomUsers[roomID] == nil {
		cm.roomUsers[roomID] = make(map[string]struct{})
	}
	cm.roomUsers[roomID][userID] = struct{}{}
	return nil
}

func (cm *memoryConnManager) LeaveRoom(_ context.Context, roomID, userID string) error {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	delete(cm.roomUsers[roomID], userID)
	return nil
}

func (cm *memoryConnManager) GetRoomMembers(_ context.Context, roomID string) ([]string, error) {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	var ids []string
	for id := range cm.roomUsers[roomID] {
		ids = append(ids, id)
	}
	return ids, nil
}

func (cm *memoryConnManager) OnlineCount() int {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	total := 0
	for _, devs := range cm.localConns {
		total += len(devs)
	}
	return total
}

func (cm *memoryConnManager) MarkHeartbeatSuccess(userID, deviceID string) {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	devs := cm.localConns[userID]
	if devs == nil {
		return
	}
	cs := devs[deviceID]
	if cs == nil {
		return
	}
	cs.LastSuccessHeartbeat = time.Now()
	cs.ConsecutiveEchoFailures = 0
}

func (cm *memoryConnManager) MarkHeartbeatFail(userID, deviceID string) {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	devs := cm.localConns[userID]
	if devs == nil {
		return
	}
	cs := devs[deviceID]
	if cs == nil {
		return
	}
	cs.ConsecutiveEchoFailures++
}

func (cm *memoryConnManager) SweepOffline(_ context.Context, timeout time.Duration) []biz.OfflineDevice {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	now := time.Now()
	var offline []biz.OfflineDevice
	for userID, devs := range cm.localConns {
		for deviceID, cs := range devs {
			if now.Sub(cs.LastSuccessHeartbeat) > timeout {
				offline = append(offline, biz.OfflineDevice{
					UserID:   userID,
					DeviceID: deviceID,
					Conn:     cs.Conn,
				})
				delete(devs, deviceID)
			}
		}
		if len(devs) == 0 {
			delete(cm.localConns, userID)
		}
	}
	return offline
}
