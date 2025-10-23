package user

import (
	"github.com/wisaitas/github.com/wisaitas/golang-structure/internal/golangstructure/domain/repository"
	"github.com/wisaitas/github.com/wisaitas/golang-structure/internal/golangstructure/usecase/user/createuser"

	"github.com/gofiber/fiber/v2"
)

type Strategy interface {
	handler
}

type handler interface {
	CreateUser(c *fiber.Ctx) error
}

type strategy struct {
	createuser createuser.Handler
}

func New(
	userRepository repository.UserRepository,
) Strategy {
	return &strategy{
		createuser: createuser.New(userRepository),
	}
}

func (s *strategy) CreateUser(c *fiber.Ctx) error {
	return s.createuser.CreateUser(c)
}
