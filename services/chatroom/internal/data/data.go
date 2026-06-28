package data

import (
	"github.com/google/wire"
	"github.com/murphy-hc/h-im/pkg/database"
	"github.com/murphy-hc/h-im/services/chatroom/internal/biz"
	"github.com/murphy-hc/h-im/services/chatroom/internal/conf"
	"gorm.io/gorm"
)

var ProviderSet = wire.NewSet(NewData, NewChatroomRepo)

// Data holds data source clients.
type Data struct {
	DB *gorm.DB
}

// NewData creates a Data instance.
func NewData(bc *conf.Bootstrap) (*Data, func(), error) {
	pg := bc.GetData().GetDatabase().GetChatroom()
	db, dbCleanup, err := database.NewDB(&database.Config{
		DSN:          pg.GetDsn(),
		MaxIdleConns: int(pg.GetMaxIdleConns()),
		MaxOpenConns: int(pg.GetMaxOpenConns()),
	})
	if err != nil {
		return nil, nil, err
	}
	return &Data{DB: db}, dbCleanup, nil
}

// Migrate runs auto-migration.
func (d *Data) Migrate() error {
	return d.DB.AutoMigrate(&RoomModel{}, &RoomMemberModel{})
}

var _ biz.ChatroomRepo = (*chatroomRepo)(nil)
