package data

import (
	"context"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/wire"
	"github.com/murphy-hc/h-im/pkg/database"
	"github.com/murphy-hc/h-im/pkg/redis"
	"github.com/murphy-hc/h-im/services/gateway/internal/biz"
	"github.com/murphy-hc/h-im/services/gateway/internal/conf"
	goredis "github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

// ProviderSet is data providers.
var ProviderSet = wire.NewSet(NewData, NewRedisClient, NewConnManager, NewAppRepo)

// Data holds data source clients.
type Data struct {
	DB *gorm.DB
}

// NewData creates a Data instance using the shared database package.
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
func (d *Data) Migrate() error {
	if err := d.DB.AutoMigrate(&AppModel{}); err != nil {
		return err
	}
	return nil
}

// NewRedisClient creates a Redis client from config.
func NewRedisClient(bc *conf.Bootstrap) (*goredis.Client, func(), error) {
	cfg := bc.GetData().GetRedis()
	addr := cfg.GetAddr()
	if addr == "" {
		addr = "localhost:6379"
	}
	rdb, err := redis.NewClient(context.Background(), redis.Config{
		Host:     addr,
		Password: cfg.GetPassword(),
	})
	if err != nil {
		return nil, nil, err
	}
	cleanup := func() {
		if err := rdb.Close(); err != nil {
			log.Errorf("redis close: %v", err)
		}
	}
	return rdb, cleanup, nil
}

func NewConnManager(rdb *goredis.Client) biz.ConnManager { return newRedisConnManager(rdb) }
func NewMemoryConnManager() biz.ConnManager               { return newMemoryConnManager() }
