package updateuser

import (
	"github.com/gofiber/fiber/v2"
	"github.com/wisaitas/github.com/wisaitas/golang-structure/internal/golangstructure/domain/entity"
	"github.com/wisaitas/github.com/wisaitas/golang-structure/internal/golangstructure/domain/repository"
)

type service interface {
	UpdateUser(c *fiber.Ctx, req *Request, id int) error
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

func (s *serviceImpl) UpdateUser(c *fiber.Ctx, req *Request, id int) error {
	user := entity.User{
		ID:   id,
		Name: req.Name,
		Age:  req.Age,
	}

	if err := s.userRepository.ReplaceUser(&user); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": err.Error(),
		})
	}

	return c.Status(fiber.StatusNoContent).JSON(nil)
}
