package initial

import (
	appRepository "github.com/wisaitas/github.com/wisaitas/golang-structure/internal/golangstructure/domain/repository"
)

type repository struct {
	userRepository appRepository.UserRepository
}

func newRepository(
	config *config,
) *repository {
	return &repository{
		userRepository: appRepository.NewUserRepository(config.postgresDB),
	}
}
