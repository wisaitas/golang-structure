package createuser

import (
	"github.com/wisaitas/github.com/wisaitas/golang-structure/internal/golangstructure/domain/repository"
)

type service interface {
	CreateUser(request *Request) error
}

type serviceImpl struct {
	userRepository repository.UserRepository
}

func newService(
	userRepository repository.UserRepository,
) service {
	return &serviceImpl{
		userRepository: userRepository,
	}
}

func (s *serviceImpl) CreateUser(request *Request) error {
	return nil
}
