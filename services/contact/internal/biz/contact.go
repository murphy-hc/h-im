package biz

// ContactUseCase handles contact business logic.
type ContactUseCase struct {
	repo ContactRepo
}

// NewContactUseCase creates a ContactUseCase.
func NewContactUseCase(repo ContactRepo) *ContactUseCase {
	return &ContactUseCase{repo: repo}
}
