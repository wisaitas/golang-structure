package deleteuser

import (
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/wisaitas/github.com/wisaitas/golang-structure/internal/golangstructure/domain/entity"
	"github.com/wisaitas/github.com/wisaitas/golang-structure/internal/golangstructure/domain/repository"
	"github.com/wisaitas/github.com/wisaitas/golang-structure/pkg/httpx"
)

type Service interface {
	Service(c fiber.Ctx, userID int) error
}

type service struct {
	userRepository repository.UserRepository
}

func newService(
	userRepository repository.UserRepository,
) Service {
	return &service{
		userRepository: userRepository,
	}
}

func (s *service) Service(c fiber.Ctx, userID int) error {
	ctx := httpx.RequestContext(c)
	db := s.userRepository.GetDB(ctx).Model(&entity.User{}).Where("id = ?", userID)
	if err := s.userRepository.WithTx(db).Update(ctx, map[string]interface{}{"deleted_at": time.Now()}); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": err.Error(),
		})
	}

	return c.Status(fiber.StatusNoContent).JSON(fiber.Map{
		"message": "user deleted successfully",
	})
}
