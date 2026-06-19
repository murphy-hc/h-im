package biz

// UserUseCase handles user business logic.
type UserUseCase struct {
	repo UserRepo
}

// NewUserUseCase creates a UserUseCase.
func NewUserUseCase(repo UserRepo) *UserUseCase {
	return &UserUseCase{repo: repo}
}
