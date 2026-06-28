package data

import (
	"context"

	"github.com/google/wire"
	"github.com/murphy-hc/h-im/pkg/database"
	"github.com/murphy-hc/h-im/pkg/redis"
	"github.com/murphy-hc/h-im/services/user/internal/biz"
	"github.com/murphy-hc/h-im/services/user/internal/conf"
	goredis "github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

var ProviderSet = wire.NewSet(NewData, NewUserRepo, NewRedisClient, NewAppRepo)

// Data holds data source clients.
type Data struct {
	DB  *gorm.DB
	RDB *goredis.Client
}

// NewData creates a Data instance from config.
func NewData(bc *conf.Bootstrap) (*Data, func(), error) {
	cfg := bc.GetData().GetDatabase().GetUser()
	db, dbCleanup, err := database.NewDB(&database.Config{
		DSN:          cfg.GetDsn(),
		MaxIdleConns: int(cfg.GetMaxIdleConns()),
		MaxOpenConns: int(cfg.GetMaxOpenConns()),
	})
	if err != nil {
		return nil, nil, err
	}
	rdb, rdbCleanup, err := NewRedisClient(bc)
	if err != nil {
		dbCleanup()
		return nil, nil, err
	}
	return &Data{DB: db, RDB: rdb}, func() {
		dbCleanup()
		rdbCleanup()
	}, nil
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
	return rdb, func() { rdb.Close() }, nil
}

// Migrate runs auto-migration.
func (d *Data) Migrate() error { return d.DB.AutoMigrate(&AppModel{}, &UserModel{}) }

var _ biz.UserRepo = (*userRepo)(nil)
