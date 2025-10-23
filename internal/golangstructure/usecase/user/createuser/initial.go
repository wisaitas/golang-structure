package createuser

import "github.com/wisaitas/github.com/wisaitas/golang-structure/internal/golangstructure/domain/repository"

func New(
	userRepository repository.UserRepository,
) Handler {
	service := newService(userRepository)
	handler := newHandler(service)

	return handler
}
