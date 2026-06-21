package data

import (
	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/wire"
	"github.com/murphy-hc/h-im/pkg/database"
	"github.com/murphy-hc/h-im/services/message/internal/conf"
	"gorm.io/gorm"
)

// ProviderSet is data providers.
var ProviderSet = wire.NewSet(NewData, NewMessageRepo, NewGatewayClient, NewUserClient)

// Data holds data source clients.
type Data struct {
	DB *gorm.DB
}

// NewData creates a Data instance using the shared database package.
func NewData(bc *conf.Bootstrap) (*Data, func(), error) {
	pg := bc.GetData().GetDatabase().GetMessage()
	db, cleanup, err := database.NewDB(&database.Config{
		DSN:          pg.GetDsn(),
		MaxIdleConns: int(pg.GetMaxIdleConns()),
		MaxOpenConns: int(pg.GetMaxOpenConns()),
	})
	if err != nil {
		return nil, nil, err
	}
	log.Infof("data: connected to database")
	return &Data{DB: db}, cleanup, nil
}

// Migrate runs auto-migration.
func (d *Data) Migrate() error {
	return d.DB.AutoMigrate(&MessageModel{})
}
