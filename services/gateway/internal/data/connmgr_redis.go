package data

import (
	"context"
	"net"
	"os"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/murphy-hc/h-im/services/gateway/internal/biz"
	"github.com/redis/go-redis/v9"
)

const redisKeyPrefix = "gw"

var instanceID = func() string {
	if v := os.Getenv("INSTANCE_ID"); v != "" {
		return v
	}
	hn, _ := os.Hostname()
	return hn
}()

// gatewayAddr returns this gateway's address for cross-instance routing.
// Set via GATEWAY_ADDR env (pod IP in K8s), defaults to instanceID:9200.
func gatewayAddr() string {
	v := os.Getenv("GATEWAY_ADDR")
	if v == "" {
		return instanceID + ":9200"
	}
	if _, _, err := net.SplitHostPort(v); err != nil {
		return v + ":9200"
	}
	return v
}

type redisConnManager struct {
	rdb        *redis.Client
	mu         sync.RWMutex
	localConns map[string]map[string]*biz.ConnState
}

func newRedisConnManager(rdb *redis.Client) *redisConnManager {
	return &redisConnManager{
		rdb:        rdb,
		localConns: make(map[string]map[string]*biz.ConnState),
	}
}

func connKey(userID, deviceID string) string { return redisKeyPrefix + ":conn:" + userID + ":" + deviceID }
func groupKey(groupID string) string         { return redisKeyPrefix + ":group:" + groupID }
func roomKey(roomID string) string           { return redisKeyPrefix + ":room:" + roomID }

func (cm *redisConnManager) Add(userID, deviceID string, conn *websocket.Conn) error {
	cm.mu.Lock()
	if cm.localConns[userID] == nil {
		cm.localConns[userID] = make(map[string]*biz.ConnState)
	}
	cm.localConns[userID][deviceID] = &biz.ConnState{
		Conn:                 conn,
		LastSuccessHeartbeat: time.Now(),
	}
	cm.mu.Unlock()
	return cm.rdb.Set(context.Background(), connKey(userID, deviceID), instanceID, 0).Err()
}

func (cm *redisConnManager) Remove(userID, deviceID string) error {
	cm.mu.Lock()
	if devs, ok := cm.localConns[userID]; ok {
		delete(devs, deviceID)
		if len(devs) == 0 {
			delete(cm.localConns, userID)
		}
	}
	cm.mu.Unlock()
	return cm.rdb.Del(context.Background(), connKey(userID, deviceID)).Err()
}

func (cm *redisConnManager) GetConns(userID string) ([]*websocket.Conn, error) {
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

func (cm *redisConnManager) KickUser(userID string) ([]*websocket.Conn, error) {
	cm.mu.Lock()
	devs := cm.localConns[userID]
	delete(cm.localConns, userID)
	cm.mu.Unlock()

	conns := make([]*websocket.Conn, 0, len(devs))
	ctx := context.Background()
	for deviceID, cs := range devs {
		cs.Conn.Close()
		conns = append(conns, cs.Conn)
		cm.rdb.Del(ctx, connKey(userID, deviceID))
	}
	return conns, nil
}

func (cm *redisConnManager) JoinGroup(groupID, userID string) error {
	return cm.rdb.SAdd(context.Background(), groupKey(groupID), userID).Err()
}
func (cm *redisConnManager) LeaveGroup(groupID, userID string) error {
	return cm.rdb.SRem(context.Background(), groupKey(groupID), userID).Err()
}
func (cm *redisConnManager) GetGroupMembers(groupID string) ([]string, error) {
	return cm.rdb.SMembers(context.Background(), groupKey(groupID)).Result()
}
func (cm *redisConnManager) JoinRoom(roomID, userID string) error {
	return cm.rdb.SAdd(context.Background(), roomKey(roomID), userID).Err()
}
func (cm *redisConnManager) LeaveRoom(roomID, userID string) error {
	return cm.rdb.SRem(context.Background(), roomKey(roomID), userID).Err()
}
func (cm *redisConnManager) GetRoomMembers(roomID string) ([]string, error) {
	return cm.rdb.SMembers(context.Background(), roomKey(roomID)).Result()
}
func (cm *redisConnManager) OnlineCount() int {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	total := 0
	for _, devs := range cm.localConns {
		total += len(devs)
	}
	return total
}

// MarkHeartbeatSuccess records a successful heartbeat.
func (cm *redisConnManager) MarkHeartbeatSuccess(userID, deviceID string) {
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

// MarkHeartbeatFail records a failed heartbeat echo.
func (cm *redisConnManager) MarkHeartbeatFail(userID, deviceID string) {
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

// SweepOffline scans all connections and returns those that have exceeded the timeout.
func (cm *redisConnManager) SweepOffline(timeout time.Duration) []biz.OfflineDevice {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	now := time.Now()
	var offline []biz.OfflineDevice
	ctx := context.Background()

	for userID, devs := range cm.localConns {
		for deviceID, cs := range devs {
			if now.Sub(cs.LastSuccessHeartbeat) > timeout {
				offline = append(offline, biz.OfflineDevice{
					UserID:   userID,
					DeviceID: deviceID,
					Conn:     cs.Conn,
				})
				delete(devs, deviceID)
				cm.rdb.Del(ctx, connKey(userID, deviceID))
			}
		}
		if len(devs) == 0 {
			delete(cm.localConns, userID)
		}
	}
	return offline
}
