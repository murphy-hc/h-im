package data

import (
	"context"

	"github.com/google/wire"
	"github.com/murphy-hc/h-im/pkg/database"
	"github.com/murphy-hc/h-im/pkg/oss"
	"github.com/murphy-hc/h-im/pkg/redis"
	"github.com/murphy-hc/h-im/services/media/internal/biz"
	"github.com/murphy-hc/h-im/services/media/internal/conf"
	goredis "github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

var ProviderSet = wire.NewSet(NewData, NewMediaRepo, NewOssClient, NewOSSStorage, wire.Bind(new(biz.Storage), new(*ossStorage)))

// Data holds data source clients.
type Data struct {
	DB  *gorm.DB
	RDB *goredis.Client
}

// NewData creates a Data instance from config.
func NewData(bc *conf.Bootstrap) (*Data, func(), error) {
	pg := bc.GetData().GetDatabase().GetMedia()
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
	if addr == "" {
		addr = "localhost:6379"
	}
	rdb, err := redis.NewClient(context.Background(), redis.Config{
		Host:     addr,
		Password: rdbCfg.GetPassword(),
	})
	if err != nil {
		dbCleanup()
		return nil, nil, err
	}
	return &Data{DB: db, RDB: rdb}, func() {
		dbCleanup()
		rdb.Close()
	}, nil
}

// Migrate runs auto-migration.
func (d *Data) Migrate() error { return d.DB.AutoMigrate(&MediaModel{}) }

// NewOssClient creates an OSS client from config.
func NewOssClient(bc *conf.Bootstrap) (*oss.Client, error) {
	oc := bc.GetData().GetOss()
	return oss.NewClient(oss.Config{
		Endpoint:        oc.GetEndpoint(),
		AccessKeyID:     oc.GetAccessKeyId(),
		AccessKeySecret: oc.GetAccessKeySecret(),
		BucketName:      oc.GetBucketName(),
	})
}
