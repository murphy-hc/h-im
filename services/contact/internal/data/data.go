package data

import (
	"context"

	"github.com/google/wire"
	"github.com/murphy-hc/h-im/pkg/database"
	"github.com/murphy-hc/h-im/pkg/redis"
	"github.com/murphy-hc/h-im/services/contact/internal/biz"
	"github.com/murphy-hc/h-im/services/contact/internal/conf"
	goredis "github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

var ProviderSet = wire.NewSet(NewData, NewContactRepo)

type Data struct {
	DB  *gorm.DB
	RDB *goredis.Client
}

func NewData(bc *conf.Bootstrap) (*Data, func(), error) {
	pg := bc.GetData().GetDatabase().GetContact()
	db, dbCleanup, err := database.NewDB(&database.Config{
		DSN:          pg.GetDsn(),
		MaxIdleConns: int(pg.GetMaxIdleConns()),
		MaxOpenConns: int(pg.GetMaxOpenConns()),
	})
	if err != nil {
		return nil, nil, err
	}
	rdbCfg := bc.GetData().GetRedis()
	addr := rdbCfg.GetAddr()
	if addr == "" { addr = "localhost:6379" }
	rdb, err := redis.NewClient(context.Background(), redis.Config{Host: addr, Password: rdbCfg.GetPassword()})
	if err != nil { dbCleanup(); return nil, nil, err }
	return &Data{DB: db, RDB: rdb}, func() { dbCleanup(); rdb.Close() }, nil
}

func (d *Data) Migrate() error {
	return d.DB.AutoMigrate(&FriendModel{}, &FriendRequestModel{})
}

var _ biz.ContactRepo = (*contactRepo)(nil)
