package data

import (
	"context"
	"net"
	"strconv"
	"time"

	"github.com/google/wire"
	"github.com/murphy-hc/h-im/pkg/database"
	"github.com/murphy-hc/h-im/pkg/redis"
	"github.com/murphy-hc/h-im/services/gateway/internal/biz"
	"github.com/murphy-hc/h-im/services/gateway/internal/conf"
	goredis "github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

// ProviderSet is data providers.
var ProviderSet = wire.NewSet(
	NewData, NewRedisClient, NewConnManager, NewPubSub,
	NewUserStatusClient, GatewayAddr,
	NewGrpcMessageClient, NewKafkaMessageClient,
	NewChatroomClient, NewGroupClient,
	wire.Bind(new(biz.Broadcaster), new(*PubSub)),
	wire.Bind(new(biz.BroadcastListener), new(*PubSub)),
)

// GatewayAddr returns this gateway's gRPC address.
func GatewayAddr() string { return gatewayAddr() }

// Data holds data source clients.
type Data struct {
	DB *gorm.DB
}

// NewData creates a Data instance from config.
func NewData(bc *conf.Bootstrap) (*Data, func(), error) {
	pg := bc.GetData().GetDatabase().GetGateway()
	db, cleanup, err := database.NewDB(&database.Config{
		DSN:          pg.GetDsn(),
		MaxIdleConns: int(pg.GetMaxIdleConns()),
		MaxOpenConns: int(pg.GetMaxOpenConns()),
	})
	if err != nil {
		return nil, nil, err
	}
	return &Data{DB: db}, cleanup, nil
}

// Migrate runs auto-migration.
func (d *Data) Migrate() error { return nil }

// NewRedisClient creates a Redis client from config.
func NewRedisClient(bc *conf.Bootstrap) (*goredis.Client, func(), error) {
	cfg := bc.GetData().GetRedis()
	host, portStr := "localhost", "6379"
	if addr := cfg.GetAddr(); addr != "" {
		host, portStr, _ = net.SplitHostPort(addr)
	}
	port, _ := strconv.Atoi(portStr)
	if port == 0 { port = 6379 }
	rdb, err := redis.NewClient(context.Background(), redis.Config{
		Host: host, Port: port, Password: cfg.GetPassword(),
	})
	if err != nil { return nil, nil, err }
	return rdb, func() { rdb.Close() }, nil
}

func NewConnManager(rdb *goredis.Client, bc *conf.Bootstrap) biz.ConnManager {
	hbTimeout := 180 * time.Second
	if bc.GetHeartbeat() != nil && bc.GetHeartbeat().GetTimeoutSeconds() > 0 {
		hbTimeout = time.Duration(bc.GetHeartbeat().GetTimeoutSeconds()) * time.Second
	}
	return newRedisConnManager(rdb, hbTimeout)
}
func NewMemoryConnManager() biz.ConnManager               { return newMemoryConnManager() }
