package biz

// MessageUseCase handles message business logic.
type MessageUseCase struct {
	repo MessageRepo
}

// NewMessageUseCase creates a MessageUseCase.
func NewMessageUseCase(repo MessageRepo) *MessageUseCase {
	return &MessageUseCase{repo: repo}
}
