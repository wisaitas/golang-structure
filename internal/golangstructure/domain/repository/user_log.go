package repository

import (
	"github.com/wisaitas/golang-structure/internal/golangstructure/domain/entity"
	"github.com/wisaitas/golang-structure/pkg/db/gormx"

	"gorm.io/gorm"
)

type UserLogRepository interface {
	gormx.BaseRepository[entity.TblUserLogs]
}

type userLogRepository struct {
	gormx.BaseRepository[entity.TblUserLogs]
}

func NewUserLogRepository(
	db *gorm.DB,
) UserLogRepository {
	return &userLogRepository{
		BaseRepository: gormx.NewBaseRepository[entity.TblUserLogs](db),
	}
}
