package initial

import (
	"github.com/wisaitas/golang-structure/internal/golangstructure/usecase/auth"
	"github.com/wisaitas/golang-structure/internal/golangstructure/usecase/user"
)

type useCase struct {
	authUseCase *auth.UseCase
	userUseCase *user.UseCase
}

func newUseCase(
	repository *repository,
	sdk *sdk,
) *useCase {
	return &useCase{
		authUseCase: auth.New(repository.userRepository, repository.userLogRepository, sdk.validator, sdk.bcrypt, sdk.logger),
		userUseCase: user.New(repository.userRepository, sdk.validator),
	}
}
