package repository

import (
	"context"
	"time"

	"gorm.io/gorm"

	"github.com/nurpe/snowops-auth/internal/model"
)

type UserSessionRepository interface {
	Create(ctx context.Context, session *model.UserSession) error
	FindByTokenHash(ctx context.Context, hash string) (*model.UserSession, error)
	Revoke(ctx context.Context, id string) error
	RevokeByUser(ctx context.Context, userID string) error
	DeleteExpired(ctx context.Context, before time.Time) error
}

type userSessionRepository struct {
	db *gorm.DB
}

func NewUserSessionRepository(db *gorm.DB) UserSessionRepository {
	return &userSessionRepository{db: db}
}

func (r *userSessionRepository) Create(ctx context.Context, session *model.UserSession) error {
	return r.db.WithContext(ctx).Create(session).Error
}

func (r *userSessionRepository) FindByTokenHash(ctx context.Context, hash string) (*model.UserSession, error) {
	var session model.UserSession
	if err := r.db.WithContext(ctx).
		Where("refresh_token_hash = ? AND revoked_at IS NULL", hash).
		First(&session).Error; err != nil {
		return nil, err
	}
	return &session, nil
}

func (r *userSessionRepository) Revoke(ctx context.Context, id string) error {
	now := time.Now()
	return r.db.WithContext(ctx).
		Model(&model.UserSession{}).
		Where("id = ?", id).
		Update("revoked_at", now).
		Error
}

func (r *userSessionRepository) RevokeByUser(ctx context.Context, userID string) error {
	now := time.Now()
	return r.db.WithContext(ctx).
		Model(&model.UserSession{}).
		Where("user_id = ?", userID).
		Update("revoked_at", now).
		Error
}

func (r *userSessionRepository) DeleteExpired(ctx context.Context, before time.Time) error {
	return r.db.WithContext(ctx).
		Where("expires_at <= ?", before).
		Delete(&model.UserSession{}).
		Error
}
