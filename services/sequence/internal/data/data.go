package data

import (
	"fmt"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/wire"
	"github.com/murphy-hc/h-im/services/sequence/internal/conf"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

// ProviderSet is data providers.
var ProviderSet = wire.NewSet(NewData, NewSequenceRepo)

// Data holds data source clients.
type Data struct {
	DB *gorm.DB
}

// NewData creates a Data instance with a GORM DB connection from config.
func NewData(bc *conf.Bootstrap) (*Data, func(), error) {
	dsn := bc.GetData().GetDatabase().GetSequence().GetDsn()
	if dsn == "" {
		return nil, nil, fmt.Errorf("data: database dsn is empty")
	}

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: gormlogger.Default.LogMode(gormlogger.Warn),
	})
	if err != nil {
		return nil, nil, fmt.Errorf("data: connect db: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, nil, fmt.Errorf("data: get sql.DB: %w", err)
	}

	maxIdle := bc.GetData().GetDatabase().GetSequence().GetMaxIdleConns()
	if maxIdle > 0 {
		sqlDB.SetMaxIdleConns(int(maxIdle))
	}
	maxOpen := bc.GetData().GetDatabase().GetSequence().GetMaxOpenConns()
	if maxOpen > 0 {
		sqlDB.SetMaxOpenConns(int(maxOpen))
	}

	if err := sqlDB.Ping(); err != nil {
		return nil, nil, fmt.Errorf("data: ping db: %w", err)
	}

	if err := db.AutoMigrate(&SequenceModel{}); err != nil {
		return nil, nil, fmt.Errorf("data: migrate: %w", err)
	}

	log.Infof("data: connected to database")

	d := &Data{DB: db}
	cleanup := func() {
		sqlDB.Close()
	}
	return d, cleanup, nil
}
