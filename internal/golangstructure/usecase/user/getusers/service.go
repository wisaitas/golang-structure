package getusers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/wisaitas/github.com/wisaitas/golang-structure/internal/golangstructure/domain/entity"
	"github.com/wisaitas/github.com/wisaitas/golang-structure/internal/golangstructure/domain/repository"
)

type service interface {
	GetUsers(c *fiber.Ctx) error
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

func (s *serviceImpl) GetUsers(c *fiber.Ctx) error {
	users := []entity.User{}
	if err := s.userRepository.GetUsers(&users); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(users)
}
