package register

import (
	"github.com/wisaitas/golang-structure/internal/golangstructure/domain/repository"
	"github.com/wisaitas/golang-structure/pkg/bcryptx"
	"github.com/wisaitas/golang-structure/pkg/logx"
	"github.com/wisaitas/golang-structure/pkg/validatorx"
)

func New(
	userRepository repository.UserRepository,
	userLogRepository repository.UserLogRepository,
	validator validatorx.Validator,
	bcrypt bcryptx.Bcrypt,
	logger logx.Logger,
) *Handler {
	service := NewService(userRepository, userLogRepository, bcrypt, logger)
	handler := newHandler(service, validator)

	return handler
}
