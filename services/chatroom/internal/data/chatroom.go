package data

import (
	"github.com/murphy-hc/h-im/services/chatroom/internal/biz"
)

type chatroomRepo struct {
	data *Data
}

// NewChatroomRepo creates a ChatroomRepo implementation.
func NewChatroomRepo(data *Data) biz.ChatroomRepo {
	return &chatroomRepo{data: data}
}
