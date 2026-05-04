package repository

import (
	"github.com/wisaitas/golang-structure/internal/golangstructure/domain/entity"
	"github.com/wisaitas/golang-structure/pkg/db/gormx"

	"gorm.io/gorm"
)

type UserRepository interface {
	gormx.BaseRepository[entity.User]
}

type userRepository struct {
	gormx.BaseRepository[entity.User]
}

func NewUserRepository(
	db *gorm.DB,
) UserRepository {
	return &userRepository{
		BaseRepository: gormx.NewBaseRepository[entity.User](db),
	}
}
