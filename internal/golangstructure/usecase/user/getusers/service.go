package getusers

import (
	"context"

	"github.com/gofiber/fiber/v3"
	"github.com/wisaitas/github.com/wisaitas/golang-structure/internal/golangstructure/domain/entity"
	"github.com/wisaitas/github.com/wisaitas/golang-structure/internal/golangstructure/domain/repository"
	"github.com/wisaitas/github.com/wisaitas/golang-structure/pkg/httpx"
)

type Service interface {
	Service(ctx context.Context, request *Request) ([]Response, error)
}

type service struct {
	operation      string
	userRepository repository.UserRepository
}

func newService(
	userRepository repository.UserRepository,
) Service {
	return &service{
		operation:      "[getusers.service]",
		userRepository: userRepository,
	}
}

func (s *service) Service(ctx context.Context, request *Request) ([]Response, error) {
	users := []entity.User{}
	if err := s.userRepository.GetUsers(ctx, &users); err != nil {
		return nil, httpx.WrapErrorWithCode(s.operation, err, fiber.StatusInternalServerError, httpx.CodeInternal)
	}

	return s.mapEntitiesToResponses(users), nil
}
