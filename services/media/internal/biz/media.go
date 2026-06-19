package biz

// MediaUseCase handles media business logic.
type MediaUseCase struct {
	repo MediaRepo
}

// NewMediaUseCase creates a MediaUseCase.
func NewMediaUseCase(repo MediaRepo) *MediaUseCase {
	return &MediaUseCase{repo: repo}
}
