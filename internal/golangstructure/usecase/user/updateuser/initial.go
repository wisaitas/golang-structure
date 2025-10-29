package updateuser

import (
	"github.com/go-playground/validator/v10"
	"github.com/wisaitas/github.com/wisaitas/golang-structure/internal/golangstructure/domain/repository"
)

func New(
	userRepository repository.UserRepository,
	validator *validator.Validate,
) Handler {
	service := newService(userRepository)
	handler := newHandler(service, validator)

	return handler
}
