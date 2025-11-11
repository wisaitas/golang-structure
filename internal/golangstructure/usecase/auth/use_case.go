package auth

import (
	"github.com/wisaitas/github.com/wisaitas/golang-structure/internal/golangstructure/domain/repository"
	"github.com/wisaitas/github.com/wisaitas/golang-structure/internal/golangstructure/usecase/auth/register"
	"github.com/wisaitas/github.com/wisaitas/golang-structure/pkg/validatorx"
)

type UseCase struct {
	Register *register.Handler
}

func New(
	userRepository repository.UserRepository,
	validator validatorx.Validator,
) *UseCase {
	return &UseCase{
		Register: register.New(userRepository, validator),
	}
}
