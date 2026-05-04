package repository

import (
	"github.com/wisaitas/golang-structure/internal/golangstructure/domain/entity"
	"github.com/wisaitas/golang-structure/pkg/db/gormx"

	"gorm.io/gorm"
)

type UserLogRepository interface {
	gormx.BaseRepository[entity.UserLog]
}

type userLogRepository struct {
	gormx.BaseRepository[entity.UserLog]
}

func NewUserLogRepository(
	db *gorm.DB,
) UserLogRepository {
	return &userLogRepository{
		BaseRepository: gormx.NewBaseRepository[entity.UserLog](db),
	}
}
