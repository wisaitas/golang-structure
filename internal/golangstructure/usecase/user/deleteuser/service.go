package deleteuser

import (
	"github.com/gofiber/fiber/v2"
	"github.com/wisaitas/github.com/wisaitas/golang-structure/internal/golangstructure/domain/repository"
)

type service interface {
	DeleteUser(c *fiber.Ctx, userID int) error
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

func (s *serviceImpl) DeleteUser(c *fiber.Ctx, userID int) error {
	if err := s.userRepository.DeleteUser(userID); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": err.Error(),
		})
	}

	return c.Status(fiber.StatusNoContent).JSON(fiber.Map{
		"message": "user deleted successfully",
	})
}
