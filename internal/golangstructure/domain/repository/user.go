package repository

import (
	"github.com/wisaitas/github.com/wisaitas/golang-structure/internal/golangstructure/domain/entity"

	"gorm.io/gorm"
)

type UserRepository interface {
	CreateUser(user *entity.User) error
	GetUsers(users *[]entity.User) error
	ReplaceUser(user *entity.User) error
}

type userRepository struct {
	db *gorm.DB
}

func NewUserRepository(
	db *gorm.DB,
) UserRepository {
	return &userRepository{
		db: db,
	}
}

func (r *userRepository) CreateUser(user *entity.User) error {
	return r.db.Create(&user).Error
}

func (r *userRepository) GetUsers(users *[]entity.User) error {
	return r.db.Find(&users).Error
}

func (r *userRepository) ReplaceUser(user *entity.User) error {
	return r.db.Updates(&user).Error
}
