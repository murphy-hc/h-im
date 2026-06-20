package database

import (
	"fmt"

	"github.com/go-kratos/kratos/v2/log"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

// Config holds database connection configuration.
type Config struct {
	DSN          string
	MaxIdleConns int
	MaxOpenConns int
}

// NewDB creates a GORM DB connection with MySQL.
func NewDB(cfg *Config) (*gorm.DB, func(), error) {
	if cfg.DSN == "" {
		return nil, nil, fmt.Errorf("database: dsn is empty")
	}
	db, err := gorm.Open(mysql.Open(cfg.DSN), &gorm.Config{
		Logger: gormlogger.Default.LogMode(gormlogger.Warn),
	})
	if err != nil {
		return nil, nil, fmt.Errorf("database: connect: %w", err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		return nil, nil, fmt.Errorf("database: get sql.DB: %w", err)
	}
	if cfg.MaxIdleConns > 0 {
		sqlDB.SetMaxIdleConns(cfg.MaxIdleConns)
	}
	if cfg.MaxOpenConns > 0 {
		sqlDB.SetMaxOpenConns(cfg.MaxOpenConns)
	}
	if err := sqlDB.Ping(); err != nil {
		return nil, nil, fmt.Errorf("database: ping: %w", err)
	}
	log.Infof("database: connected")
	cleanup := func() { sqlDB.Close() }
	return db, cleanup, nil
}
