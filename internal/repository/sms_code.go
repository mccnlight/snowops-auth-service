package repository

import (
	"context"
	"time"

	"gorm.io/gorm"

	"github.com/nurpe/snowops-auth/internal/model"
)

type SmsCodeRepository interface {
	Create(ctx context.Context, code *model.SmsCode) error
	FindLatest(ctx context.Context, phone string) (*model.SmsCode, error)
	CountActiveInRange(ctx context.Context, phone string, from time.Time) (int64, error)
	MarkUsed(ctx context.Context, id string) error
}

type smsCodeRepository struct {
	db *gorm.DB
}

func NewSmsCodeRepository(db *gorm.DB) SmsCodeRepository {
	return &smsCodeRepository{db: db}
}

func (r *smsCodeRepository) Create(ctx context.Context, code *model.SmsCode) error {
	return r.db.WithContext(ctx).Create(code).Error
}

func (r *smsCodeRepository) FindLatest(ctx context.Context, phone string) (*model.SmsCode, error) {
	var smsCode model.SmsCode
	err := r.db.WithContext(ctx).
		Where("phone = ?", phone).
		Order("created_at DESC").
		First(&smsCode).Error
	if err != nil {
		return nil, err
	}
	return &smsCode, nil
}

func (r *smsCodeRepository) CountActiveInRange(ctx context.Context, phone string, from time.Time) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&model.SmsCode{}).
		Where("phone = ? AND created_at >= ?", phone, from).
		Count(&count).Error
	return count, err
}

func (r *smsCodeRepository) MarkUsed(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).
		Model(&model.SmsCode{}).
		Where("id = ?", id).
		Update("is_used", true).
		Error
}
