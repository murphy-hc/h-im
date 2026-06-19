package biz

// AuthUseCase handles auth business logic.
type AuthUseCase struct {
	repo AuthRepo
}

// NewAuthUseCase creates a AuthUseCase.
func NewAuthUseCase(repo AuthRepo) *AuthUseCase {
	return &AuthUseCase{repo: repo}
}
