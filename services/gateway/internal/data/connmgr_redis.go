package data

import (
	"context"
	"net"
	"os"
	"sync"
	"time"

	"github.com/coder/websocket"
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
	ttl        time.Duration
}

func newRedisConnManager(rdb *redis.Client, heartbeatTimeout time.Duration) *redisConnManager {
	return &redisConnManager{
		rdb:        rdb,
		localConns: make(map[string]map[string]*biz.ConnState),
		ttl:        heartbeatTimeout + 60*time.Second,
	}
}

func connKey(userID, deviceID string) string { return redisKeyPrefix + ":conn:" + userID + ":" + deviceID }
func groupKey(groupID string) string         { return redisKeyPrefix + ":group:" + groupID }
func roomKey(roomID string) string           { return redisKeyPrefix + ":room:" + roomID }

func (cm *redisConnManager) Add(ctx context.Context, userID, deviceID string, conn *websocket.Conn) error {
	cm.mu.Lock()
	if cm.localConns[userID] == nil {
		cm.localConns[userID] = make(map[string]*biz.ConnState)
	}
	cm.localConns[userID][deviceID] = &biz.ConnState{
		Conn:                 conn,
		LastSuccessHeartbeat: time.Now(),
	}
	cm.mu.Unlock()
	return cm.rdb.Set(ctx, connKey(userID, deviceID), instanceID, cm.ttl).Err()
}

func (cm *redisConnManager) Remove(ctx context.Context, userID, deviceID string) error {
	cm.mu.Lock()
	if devs, ok := cm.localConns[userID]; ok {
		delete(devs, deviceID)
		if len(devs) == 0 {
			delete(cm.localConns, userID)
		}
	}
	cm.mu.Unlock()
	return cm.rdb.Del(ctx, connKey(userID, deviceID)).Err()
}

func (cm *redisConnManager) GetConns(_ context.Context, userID string) ([]*websocket.Conn, error) {
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

func (cm *redisConnManager) KickUser(ctx context.Context, userID string) ([]*websocket.Conn, error) {
	cm.mu.Lock()
	devs := cm.localConns[userID]
	delete(cm.localConns, userID)
	cm.mu.Unlock()

	conns := make([]*websocket.Conn, 0, len(devs))
	for deviceID, cs := range devs {
		conns = append(conns, cs.Conn)
		cm.rdb.Del(ctx, connKey(userID, deviceID))
	}
	return conns, nil
}

func (cm *redisConnManager) JoinGroup(ctx context.Context, groupID, userID string) error {
	return cm.rdb.SAdd(ctx, groupKey(groupID), userID).Err()
}
func (cm *redisConnManager) LeaveGroup(ctx context.Context, groupID, userID string) error {
	return cm.rdb.SRem(ctx, groupKey(groupID), userID).Err()
}
func (cm *redisConnManager) GetGroupMembers(ctx context.Context, groupID string) ([]string, error) {
	return cm.rdb.SMembers(ctx, groupKey(groupID)).Result()
}
func (cm *redisConnManager) JoinRoom(ctx context.Context, roomID, userID string) error {
	return cm.rdb.SAdd(ctx, roomKey(roomID), userID).Err()
}
func (cm *redisConnManager) LeaveRoom(ctx context.Context, roomID, userID string) error {
	return cm.rdb.SRem(ctx, roomKey(roomID), userID).Err()
}
func (cm *redisConnManager) GetRoomMembers(ctx context.Context, roomID string) ([]string, error) {
	return cm.rdb.SMembers(ctx, roomKey(roomID)).Result()
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
func (cm *redisConnManager) SweepOffline(ctx context.Context, timeout time.Duration) []biz.OfflineDevice {
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
				cm.rdb.Del(ctx, connKey(userID, deviceID))
			}
		}
		if len(devs) == 0 {
			delete(cm.localConns, userID)
		}
	}
	return offline
}
