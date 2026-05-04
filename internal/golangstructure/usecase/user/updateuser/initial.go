package updateuser

import (
	"github.com/wisaitas/golang-structure/internal/golangstructure/domain/repository"
	"github.com/wisaitas/golang-structure/pkg/validatorx"
)

func New(
	userRepository repository.UserRepository,
	validator validatorx.Validator,
) *Handler {
	service := newService(userRepository)
	handler := newHandler(service, validator)

	return handler
}
