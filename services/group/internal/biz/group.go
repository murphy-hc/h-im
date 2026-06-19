package biz

// GroupUseCase handles group business logic.
type GroupUseCase struct {
	repo GroupRepo
}

// NewGroupUseCase creates a GroupUseCase.
func NewGroupUseCase(repo GroupRepo) *GroupUseCase {
	return &GroupUseCase{repo: repo}
}
