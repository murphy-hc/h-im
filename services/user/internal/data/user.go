package data

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/murphy-hc/h-im/services/user/internal/biz"
	"github.com/murphy-hc/h-im/services/user/internal/conf"
	"github.com/redis/go-redis/v9"
)

type userRepo struct {
	data *Data
	ttl  time.Duration
}

// NewUserRepo creates a UserRepo implementation. Redis keys expire at heartbeat_timeout + 60s.
func NewUserRepo(data *Data, bc *conf.Bootstrap) biz.UserRepo {
	hbTimeout := 180 * time.Second
	if bc.GetHeartbeat() != nil && bc.GetHeartbeat().GetTimeoutSeconds() > 0 {
		hbTimeout = time.Duration(bc.GetHeartbeat().GetTimeoutSeconds()) * time.Second
	}
	return &userRepo{data: data, ttl: hbTimeout + 60*time.Second}
}

func onlineKey(userID, deviceID string) string { return fmt.Sprintf("user:online:%s:%s", userID, deviceID) }
func devicesKey(userID string) string          { return fmt.Sprintf("user:devices:%s", userID) }

func (r *userRepo) SetOnline(ctx context.Context, userID, deviceID, gatewayAddr string, timestamp int64) error {
	key := onlineKey(userID, deviceID)
	pipe := r.data.RDB.Pipeline()
	pipe.HSet(ctx, key,
		"gateway_addr", gatewayAddr,
		"last_heartbeat_ts", timestamp,
	)
	pipe.Expire(ctx, key, r.ttl)
	pipe.SAdd(ctx, devicesKey(userID), deviceID)
	pipe.Expire(ctx, devicesKey(userID), r.ttl)
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
// Uses LastIndex so that user IDs containing colons are handled correctly.
func splitOnlineKey(key string) []string {
	prefixLen := len("user:online:")
	if len(key) <= prefixLen {
		return nil
	}
	rest := key[prefixLen:]
	if idx := strings.LastIndex(rest, ":"); idx >= 0 {
		return []string{rest[:idx], rest[idx+1:]}
	}
	return nil
}

var _ redis.Cmdable = (*redis.Client)(nil)

func (r *userRepo) Register(ctx context.Context, userID, username, passwordHash string) error {
	return r.data.DB.WithContext(ctx).Create(&UserModel{
		UserID: userID, Username: username, PasswordHash: passwordHash,
	}).Error
}

func (r *userRepo) FindByUsername(ctx context.Context, username string) (string, string, error) {
	var m UserModel
	if err := r.data.DB.WithContext(ctx).Where("username = ?", username).First(&m).Error; err != nil {
		return "", "", err
	}
	return m.UserID, m.PasswordHash, nil
}

func (r *userRepo) FindByUserID(ctx context.Context, userID string) (*biz.User, error) {
	var m UserModel
	if err := r.data.DB.WithContext(ctx).Where("user_id = ?", userID).First(&m).Error; err != nil {
		return nil, err
	}
	return &biz.User{
		UserID: m.UserID, Username: m.Username, Nickname: m.Nickname, Avatar: m.Avatar,
	}, nil
}

func (r *userRepo) BatchGetUsers(ctx context.Context, userIDs []string) ([]*biz.User, error) {
	if len(userIDs) == 0 {
		return nil, nil
	}
	var models []UserModel
	if err := r.data.DB.WithContext(ctx).Where("user_id IN ?", userIDs).Find(&models).Error; err != nil {
		return nil, err
	}
	result := make([]*biz.User, 0, len(models))
	for i := range models {
		result = append(result, &biz.User{
			UserID: models[i].UserID, Username: models[i].Username,
			Nickname: models[i].Nickname, Avatar: models[i].Avatar,
		})
	}
	return result, nil
}

func (r *userRepo) UpdateProfile(ctx context.Context, userID, nickname, avatar string) error {
	updates := map[string]any{}
	if nickname != "" {
		updates["nickname"] = nickname
	}
	if avatar != "" {
		updates["avatar"] = avatar
	}
	if len(updates) == 0 {
		return nil
	}
	return r.data.DB.WithContext(ctx).Model(&UserModel{}).Where("user_id = ?", userID).Updates(updates).Error
}

func (r *userRepo) FindAppByID(ctx context.Context, appID string) (*biz.App, error) {
	var model AppModel
	err := r.data.DB.WithContext(ctx).Where("app_id = ? AND enabled = true", appID).First(&model).Error
	if err != nil {
		return nil, fmt.Errorf("find app %s: %w", appID, err)
	}
	return &biz.App{
		AppID: model.AppID, AppSecret: model.AppSecret,
		AppName: model.AppName, Enabled: model.Enabled,
	}, nil
}
