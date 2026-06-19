package biz

// ChatroomUseCase handles chatroom business logic.
type ChatroomUseCase struct {
	repo ChatroomRepo
}

// NewChatroomUseCase creates a ChatroomUseCase.
func NewChatroomUseCase(repo ChatroomRepo) *ChatroomUseCase {
	return &ChatroomUseCase{repo: repo}
}
