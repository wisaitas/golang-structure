package user

import (
	"github.com/go-playground/validator/v10"
	"github.com/wisaitas/github.com/wisaitas/golang-structure/internal/golangstructure/domain/repository"
	"github.com/wisaitas/github.com/wisaitas/golang-structure/internal/golangstructure/usecase/user/createuser"
	"github.com/wisaitas/github.com/wisaitas/golang-structure/internal/golangstructure/usecase/user/deleteuser"
	"github.com/wisaitas/github.com/wisaitas/golang-structure/internal/golangstructure/usecase/user/getusers"
	"github.com/wisaitas/github.com/wisaitas/golang-structure/internal/golangstructure/usecase/user/updateuser"

	"github.com/gofiber/fiber/v2"
)

type Strategy interface {
	handler
}

type handler interface {
	CreateUser(c *fiber.Ctx) error
	GetUsers(c *fiber.Ctx) error
	UpdateUser(c *fiber.Ctx) error
	DeleteUser(c *fiber.Ctx) error
}

type strategy struct {
	createuser createuser.Handler
	getusers   getusers.Handler
	updateuser updateuser.Handler
	deleteuser deleteuser.Handler
}

func New(
	userRepository repository.UserRepository,
	validator *validator.Validate,
) Strategy {
	return &strategy{
		createuser: createuser.New(userRepository),
		getusers:   getusers.New(userRepository),
		updateuser: updateuser.New(userRepository, validator),
		deleteuser: deleteuser.New(userRepository),
	}
}

func (s *strategy) CreateUser(c *fiber.Ctx) error {
	return s.createuser.CreateUser(c)
}

func (s *strategy) GetUsers(c *fiber.Ctx) error {
	return s.getusers.GetUsers(c)
}

func (s *strategy) UpdateUser(c *fiber.Ctx) error {
	return s.updateuser.UpdateUser(c)
}

func (s *strategy) DeleteUser(c *fiber.Ctx) error {
	return s.deleteuser.DeleteUser(c)
}
