package repository

import (
	"context"
	"time"

	"github.com/wisaitas/github.com/wisaitas/golang-structure/internal/golangstructure/domain/entity"
	"github.com/wisaitas/github.com/wisaitas/golang-structure/pkg/httpx"

	"gorm.io/gorm"
)

type UserRepository interface {
	CreateUser(ctx context.Context, user *entity.User) error
	GetUsers(ctx context.Context, users *[]entity.User) error
	ReplaceUser(ctx context.Context, user *entity.User) error
	DeleteUser(ctx context.Context, userID int) error
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

func (r *userRepository) CreateUser(ctx context.Context, user *entity.User) error {
	if err := r.db.WithContext(ctx).Create(&user).Error; err != nil {
		return httpx.WrapError("register.repo.create_user", err, 0)
	}
	return nil
}

func (r *userRepository) GetUsers(ctx context.Context, users *[]entity.User) error {
	return r.db.WithContext(ctx).Where("deleted_at IS NULL").Find(&users).Error
}

func (r *userRepository) ReplaceUser(ctx context.Context, user *entity.User) error {
	return r.db.WithContext(ctx).Updates(&user).Error
}

func (r *userRepository) DeleteUser(ctx context.Context, userID int) error {
	return r.db.WithContext(ctx).Model(&entity.User{}).Where("id = ?", userID).Updates(map[string]interface{}{"deleted_at": time.Now()}).Error
}
