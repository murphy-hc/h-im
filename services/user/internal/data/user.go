package data

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/murphy-hc/h-im/services/user/internal/biz"
	"github.com/redis/go-redis/v9"
)

type userRepo struct {
	data *Data
}

// NewUserRepo creates a UserRepo implementation.
func NewUserRepo(data *Data) biz.UserRepo {
	return &userRepo{data: data}
}

func onlineKey(userID, deviceID string) string { return fmt.Sprintf("user:online:%s:%s", userID, deviceID) }
func devicesKey(userID string) string          { return fmt.Sprintf("user:devices:%s", userID) }

func (r *userRepo) SetOnline(ctx context.Context, userID, deviceID, gatewayAddr string, timestamp int64) error {
	pipe := r.data.RDB.Pipeline()
	pipe.HSet(ctx, onlineKey(userID, deviceID),
		"gateway_addr", gatewayAddr,
		"last_heartbeat_ts", timestamp,
	)
	pipe.SAdd(ctx, devicesKey(userID), deviceID)
	_, err := pipe.Exec(ctx)
	return err
}

func (r *userRepo) SetOffline(ctx context.Context, userID, deviceID string) error {
	pipe := r.data.RDB.Pipeline()
	pipe.Del(ctx, onlineKey(userID, deviceID))
	pipe.SRem(ctx, devicesKey(userID), deviceID)
	_, err := pipe.Exec(ctx)
	return err
}

func (r *userRepo) GetUserOnline(ctx context.Context, userID string) ([]biz.OnlineDevice, error) {
	deviceIDs, err := r.data.RDB.SMembers(ctx, devicesKey(userID)).Result()
	if err != nil {
		return nil, err
	}
	if len(deviceIDs) == 0 {
		return nil, nil
	}
	result := make([]biz.OnlineDevice, 0, len(deviceIDs))
	for _, deviceID := range deviceIDs {
		fields, err := r.data.RDB.HGetAll(ctx, onlineKey(userID, deviceID)).Result()
		if err != nil {
			continue
		}
		ts, _ := strconv.ParseInt(fields["last_heartbeat_ts"], 10, 64)
		result = append(result, biz.OnlineDevice{
			DeviceID:      deviceID,
			GatewayAddr:   fields["gateway_addr"],
			LastHeartbeat: ts,
		})
	}
	return result, nil
}

// SweepOffline scans Redis for devices with stale heartbeats and removes them.
func (r *userRepo) SweepOffline(ctx context.Context, timeoutSeconds int64) ([]biz.OfflinePair, error) {
	now := time.Now()
	// SCAN for all online keys
	var offline []biz.OfflinePair
	var cursor uint64
	pattern := "user:online:*"
	cutoff := int64(0) // placeholder — we compare per-key
	_ = cutoff

	for {
		keys, nextCursor, err := r.data.RDB.Scan(ctx, cursor, pattern, 100).Result()
		if err != nil {
			return nil, err
		}
		for _, key := range keys {
			tsStr, err := r.data.RDB.HGet(ctx, key, "last_heartbeat_ts").Result()
			if err != nil {
				continue
			}
			ts, err := strconv.ParseInt(tsStr, 10, 64)
			if err != nil {
				continue
			}
			// key format: user:online:{userID}:{deviceID}
			parts := splitOnlineKey(key)
			if len(parts) != 2 {
				continue
			}
			userID, deviceID := parts[0], parts[1]
			if now.Unix()-ts > timeoutSeconds {
				pipe := r.data.RDB.Pipeline()
				pipe.Del(ctx, key)
				pipe.SRem(ctx, devicesKey(userID), deviceID)
				pipe.Exec(ctx)
				offline = append(offline, biz.OfflinePair{UserID: userID, DeviceID: deviceID})
			}
		}
		cursor = nextCursor
		if cursor == 0 {
			break
		}
	}
	return offline, nil
}

// splitOnlineKey parses "user:online:{userID}:{deviceID}" into [userID, deviceID].
func splitOnlineKey(key string) []string {
	// key = "user:online:user123:device456"
	prefixLen := len("user:online:")
	if len(key) <= prefixLen {
		return nil
	}
	rest := key[prefixLen:]
	for i := 0; i < len(rest); i++ {
		if rest[i] == ':' {
			return []string{rest[:i], rest[i+1:]}
		}
	}
	return nil
}

var _ redis.Cmdable = (*redis.Client)(nil)
