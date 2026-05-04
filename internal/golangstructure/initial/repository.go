package initial

import (
	appRepository "github.com/wisaitas/golang-structure/internal/golangstructure/domain/repository"
)

type repository struct {
	userRepository    appRepository.UserRepository
	userLogRepository appRepository.UserLogRepository
}

func newRepository(
	config *config,
) *repository {
	return &repository{
		userRepository:    appRepository.NewUserRepository(config.sqlDB),
		userLogRepository: appRepository.NewUserLogRepository(config.sqlDB),
	}
}
