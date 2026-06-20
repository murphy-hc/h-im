package data

import (
	"context"
	"os"
	"sync"

	"github.com/gorilla/websocket"
	"github.com/redis/go-redis/v9"
)

const redisKeyPrefix = "gw"

var instanceID = func() string {
	if v := os.Getenv("INSTANCE_ID"); v != "" { return v }
	hn, _ := os.Hostname()
	return hn
}()

type redisConnManager struct {
	rdb        *redis.Client
	mu         sync.RWMutex
	localConns map[string]map[string]*websocket.Conn
}

func newRedisConnManager(rdb *redis.Client) *redisConnManager {
	return &redisConnManager{rdb: rdb, localConns: make(map[string]map[string]*websocket.Conn)}
}

func connKey(userID, deviceID string) string { return redisKeyPrefix + ":conn:" + userID + ":" + deviceID }
func groupKey(groupID string) string         { return redisKeyPrefix + ":group:" + groupID }
func roomKey(roomID string) string           { return redisKeyPrefix + ":room:" + roomID }

func (cm *redisConnManager) Add(userID, deviceID string, conn *websocket.Conn) error {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	if cm.localConns[userID] == nil {
		cm.localConns[userID] = make(map[string]*websocket.Conn)
	}
	cm.localConns[userID][deviceID] = conn
	return cm.rdb.Set(context.Background(), connKey(userID, deviceID), instanceID, 0).Err()
}

func (cm *redisConnManager) Remove(userID, deviceID string) error {
	cm.mu.Lock()
	if devs, ok := cm.localConns[userID]; ok {
		delete(devs, deviceID)
		if len(devs) == 0 { delete(cm.localConns, userID) }
	}
	cm.mu.Unlock()
	return cm.rdb.Del(context.Background(), connKey(userID, deviceID)).Err()
}

func (cm *redisConnManager) GetConns(userID string) ([]*websocket.Conn, error) {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	devs := cm.localConns[userID]
	if len(devs) == 0 { return nil, nil }
	conns := make([]*websocket.Conn, 0, len(devs))
	for _, c := range devs { conns = append(conns, c) }
	return conns, nil
}

func (cm *redisConnManager) KickUser(userID string) ([]*websocket.Conn, error) {
	cm.mu.Lock()
	devs := cm.localConns[userID]
	delete(cm.localConns, userID)
	cm.mu.Unlock()
	conns := make([]*websocket.Conn, 0, len(devs))
	for deviceID, conn := range devs {
		conn.Close()
		conns = append(conns, conn)
		cm.rdb.Del(context.Background(), connKey(userID, deviceID))
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
	cm.mu.RLock(); defer cm.mu.RUnlock()
	total := 0
	for _, devs := range cm.localConns { total += len(devs) }
	return total
}
