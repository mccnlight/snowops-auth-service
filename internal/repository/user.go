package repository

import (
	"context"

	"gorm.io/gorm"

	"github.com/nurpe/snowops-auth/internal/model"
)

type UserRepository interface {
	FindByLogin(ctx context.Context, login string) (*model.User, error)
	FindByPhone(ctx context.Context, phone string) (*model.User, error)
	FindByID(ctx context.Context, id string) (*model.User, error)
	Create(ctx context.Context, user *model.User) error
	ExistsByLogin(ctx context.Context, login string) (bool, error)
	ExistsByPhone(ctx context.Context, phone string) (bool, error)
}

type userRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) FindByLogin(ctx context.Context, login string) (*model.User, error) {
	var user model.User
	if err := r.db.WithContext(ctx).
		Where("login = ? AND is_active = TRUE", login).
		Preload("Organization").
		First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) FindByPhone(ctx context.Context, phone string) (*model.User, error) {
	var user model.User
	if err := r.db.WithContext(ctx).
		Where("phone = ? AND is_active = TRUE", phone).
		Preload("Organization").
		First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) FindByID(ctx context.Context, id string) (*model.User, error) {
	var user model.User
	if err := r.db.WithContext(ctx).
		Where("id = ?", id).
		Preload("Organization").
		First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) Create(ctx context.Context, user *model.User) error {
	return r.db.WithContext(ctx).Create(user).Error
}

func (r *userRepository) ExistsByLogin(ctx context.Context, login string) (bool, error) {
	var count int64
	if err := r.db.WithContext(ctx).
		Model(&model.User{}).
		Where("login = ?", login).
		Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *userRepository) ExistsByPhone(ctx context.Context, phone string) (bool, error) {
	var count int64
	if err := r.db.WithContext(ctx).
		Model(&model.User{}).
		Where("phone = ?", phone).
		Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}
