package biz

// Message is the domain entity for a private message.
type Message struct {
	ServerID   int64
	ClientID   string
	SenderID   string
	ReceiverID string
	ConvType   int32
	MsgType    int32
	Text       string
	ServerTime int64
	CreateTime int64
	IsDeleted  bool
	IsRead     bool
}
