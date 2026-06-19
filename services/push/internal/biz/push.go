package biz

// PushUseCase handles push business logic.
type PushUseCase struct {
	repo PushRepo
}

// NewPushUseCase creates a PushUseCase.
func NewPushUseCase(repo PushRepo) *PushUseCase {
	return &PushUseCase{repo: repo}
}
