package initial

import (
	"github.com/go-playground/validator/v10"
	"github.com/wisaitas/github.com/wisaitas/golang-structure/internal/golangstructure/usecase/user"
)

type strategy struct {
	userStrategy user.Strategy
}

func newStrategy(
	repository *repository,
	validator *validator.Validate,
) *strategy {
	return &strategy{
		userStrategy: user.New(repository.userRepository, validator),
	}
}
