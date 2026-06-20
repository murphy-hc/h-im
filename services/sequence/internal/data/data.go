package data

import (
	"fmt"

	"github.com/google/wire"
	"github.com/murphy-hc/h-im/pkg/database"
	"github.com/murphy-hc/h-im/services/sequence/internal/conf"
	"gorm.io/gorm"
)

// ProviderSet is data providers.
var ProviderSet = wire.NewSet(NewData, NewSequenceRepo)

// Data holds data source clients.
type Data struct {
	DB *gorm.DB
}

// NewData creates a Data instance using the shared database package.
func NewData(bc *conf.Bootstrap) (*Data, func(), error) {
	pg := bc.GetData().GetDatabase().GetSequence()
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
	if err := d.DB.AutoMigrate(&SequenceModel{}); err != nil {
		return fmt.Errorf("data: migrate: %w", err)
	}
	return nil
}
