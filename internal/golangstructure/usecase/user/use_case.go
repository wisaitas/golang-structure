package user

import (
	"github.com/wisaitas/golang-structure/internal/golangstructure/domain/repository"
	"github.com/wisaitas/golang-structure/internal/golangstructure/usecase/user/createuser"
	"github.com/wisaitas/golang-structure/internal/golangstructure/usecase/user/deleteuser"
	"github.com/wisaitas/golang-structure/internal/golangstructure/usecase/user/getusers"
	"github.com/wisaitas/golang-structure/internal/golangstructure/usecase/user/updateuser"
	"github.com/wisaitas/golang-structure/pkg/validatorx"
)

type UseCase struct {
	CreateUser *createuser.Handler
	GetUsers   *getusers.Handler
	UpdateUser *updateuser.Handler
	DeleteUser *deleteuser.Handler
}

func New(
	userRepository repository.UserRepository,
	validator validatorx.Validator,
) *UseCase {
	return &UseCase{
		CreateUser: createuser.New(userRepository),
		GetUsers:   getusers.New(userRepository, validator),
		UpdateUser: updateuser.New(userRepository, validator),
		DeleteUser: deleteuser.New(userRepository),
	}
}
