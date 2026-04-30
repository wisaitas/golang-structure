package register

import (
	"context"
	"strings"

	"github.com/gofiber/fiber/v3"
	"github.com/wisaitas/github.com/wisaitas/golang-structure/internal/golangstructure"
	"github.com/wisaitas/github.com/wisaitas/golang-structure/internal/golangstructure/domain/entity"
	"github.com/wisaitas/github.com/wisaitas/golang-structure/internal/golangstructure/domain/repository"
	"github.com/wisaitas/github.com/wisaitas/golang-structure/pkg/bcryptx"
	"github.com/wisaitas/github.com/wisaitas/golang-structure/pkg/db/gormx"
	"github.com/wisaitas/github.com/wisaitas/golang-structure/pkg/httpx"
)

type Service interface {
	Service(ctx context.Context, request *Request) error
}

type service struct {
	operation         string
	userRepository    repository.UserRepository
	userLogRepository repository.UserLogRepository
	bcrypt            bcryptx.Bcrypt
}

func NewService(
	userRepository repository.UserRepository,
	userLogRepository repository.UserLogRepository,
	bcrypt bcryptx.Bcrypt,
) Service {
	return &service{
		operation:         "[register.service]",
		userRepository:    userRepository,
		userLogRepository: userLogRepository,
		bcrypt:            bcrypt,
	}
}

func (s *service) Service(ctx context.Context, request *Request) error {
	user := s.mapRequestToEntity(request)

	hashedPassword, err := s.bcrypt.GenerateFromPassword(user.Password, golangstructure.Config.Bcrypt.Cost)
	if err != nil {
		return httpx.WrapErrorWithCode(s.operation, err, fiber.StatusInternalServerError, httpx.CodeInternal)
	}

	user.Password = string(hashedPassword)

	if err := s.userRepository.Transaction(ctx, func(txRepo gormx.BaseRepository[entity.User]) error {
		if err := txRepo.Create(ctx, user); err != nil {
			if strings.Contains(err.Error(), "duplicate key") {
				return httpx.WrapErrorWithCode(s.operation, err, fiber.StatusConflict, httpx.CodeConflict)
			}

			return httpx.WrapErrorWithCode(s.operation, err, fiber.StatusInternalServerError, httpx.CodeInternal)
		}

		userLog := s.mapRequestToUserLog(user)
		if err := s.userLogRepository.WithTx(txRepo.GetDB(ctx)).Create(ctx, userLog); err != nil {
			return httpx.WrapErrorWithCode(s.operation, err, fiber.StatusInternalServerError, httpx.CodeInternal)
		}

		return nil
	}); err != nil {
		if httpx.StatusCodeFromError(err, 0) > 0 {
			return err
		}
		return httpx.WrapErrorWithCode(s.operation, err, fiber.StatusInternalServerError, httpx.CodeInternal)
	}

	return nil
}
