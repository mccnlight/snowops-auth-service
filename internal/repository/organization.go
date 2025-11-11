package repository

import (
	"context"

	"gorm.io/gorm"

	"github.com/nurpe/snowops-auth/internal/model"
)

type OrganizationRepository interface {
	FindByName(ctx context.Context, name string) (*model.Organization, error)
	FindByID(ctx context.Context, id string) (*model.Organization, error)
	ListByParent(ctx context.Context, parentID string) ([]model.Organization, error)
	Create(ctx context.Context, organization *model.Organization) error
}

type organizationRepository struct {
	db *gorm.DB
}

func NewOrganizationRepository(db *gorm.DB) OrganizationRepository {
	return &organizationRepository{db: db}
}

func (r *organizationRepository) FindByName(ctx context.Context, name string) (*model.Organization, error) {
	var org model.Organization
	if err := r.db.WithContext(ctx).Where("name = ?", name).First(&org).Error; err != nil {
		return nil, err
	}
	return &org, nil
}

func (r *organizationRepository) FindByID(ctx context.Context, id string) (*model.Organization, error) {
	var org model.Organization
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&org).Error; err != nil {
		return nil, err
	}
	return &org, nil
}

func (r *organizationRepository) ListByParent(ctx context.Context, parentID string) ([]model.Organization, error) {
	var orgs []model.Organization
	if err := r.db.WithContext(ctx).
		Where("parent_org_id = ?", parentID).
		Find(&orgs).Error; err != nil {
		return nil, err
	}
	return orgs, nil
}

func (r *organizationRepository) Create(ctx context.Context, organization *model.Organization) error {
	return r.db.WithContext(ctx).Create(organization).Error
}
