package register

import (
	"github.com/wisaitas/github.com/wisaitas/golang-structure/internal/golangstructure/domain/repository"
	"github.com/wisaitas/github.com/wisaitas/golang-structure/pkg/bcryptx"
	"github.com/wisaitas/github.com/wisaitas/golang-structure/pkg/validatorx"
)

func New(
	userRepository repository.UserRepository,
	validator validatorx.Validator,
	bcrypt bcryptx.Bcrypt,
) *Handler {
	service := newService(userRepository, bcrypt)
	handler := newHandler(service, validator)

	return handler
}
