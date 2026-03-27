package register

import (
	"strings"

	"github.com/gofiber/fiber/v3"
	"github.com/wisaitas/github.com/wisaitas/golang-structure/internal/golangstructure"
	"github.com/wisaitas/github.com/wisaitas/golang-structure/internal/golangstructure/domain/repository"
	"github.com/wisaitas/github.com/wisaitas/golang-structure/pkg/bcryptx"
	"github.com/wisaitas/github.com/wisaitas/golang-structure/pkg/httpx"
)

type Service interface {
	Service(request *Request) error
}

type service struct {
	userRepository repository.UserRepository
	bcrypt         bcryptx.Bcrypt
}

func newService(
	userRepository repository.UserRepository,
	bcrypt bcryptx.Bcrypt,
) Service {
	return &service{
		userRepository: userRepository,
		bcrypt:         bcrypt,
	}
}

func (s *service) Service(request *Request) error {
	user := s.mapRequestToEntity(request)

	hashedPassword, err := s.bcrypt.GenerateFromPassword(user.Password, golangstructure.Config.Bcrypt.Cost)
	if err != nil {
		return httpx.WrapError("register.service.hash_password", err, fiber.StatusInternalServerError)
	}

	user.Password = string(hashedPassword)

	if err := s.userRepository.CreateUser(user); err != nil {
		if strings.Contains(err.Error(), "duplicate key") {
			return httpx.WrapError("register.service.create_user", err, fiber.StatusConflict)
		}

		return httpx.WrapError("register.service.create_user", err, fiber.StatusInternalServerError)
	}

	return nil
}
