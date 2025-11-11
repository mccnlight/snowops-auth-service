package service

import (
	"context"
	"errors"
	"strings"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/nurpe/snowops-auth/internal/adapter/password"
	"github.com/nurpe/snowops-auth/internal/model"
	"github.com/nurpe/snowops-auth/internal/repository"
)

type AdminService struct {
	users         repository.UserRepository
	organizations repository.OrganizationRepository
	password      password.Hasher
}

func NewAdminService(
	users repository.UserRepository,
	organizations repository.OrganizationRepository,
	password password.Hasher,
) *AdminService {
	return &AdminService{
		users:         users,
		organizations: organizations,
		password:      password,
	}
}

type CreateOrganizationInput struct {
	Name         string
	BIN          string
	HeadFullName string
	Address      string
	Phone        string
	Admin        CreateOrganizationAdminInput
}

type CreateOrganizationAdminInput struct {
	Login    *string
	Password *string
	Phone    *string
}

type OrganizationInfo struct {
	ID           uuid.UUID              `json:"id"`
	ParentOrgID  *uuid.UUID             `json:"parent_org_id,omitempty"`
	Type         model.OrganizationType `json:"type"`
	Name         string                 `json:"name"`
	BIN          string                 `json:"bin,omitempty"`
	HeadFullName string                 `json:"head_full_name,omitempty"`
	Address      string                 `json:"address,omitempty"`
	Phone        string                 `json:"phone,omitempty"`
	IsActive     bool                   `json:"is_active"`
}

type CreateOrganizationResult struct {
	Organization OrganizationInfo `json:"organization"`
	Admin        UserInfo         `json:"admin"`
}

type CreateUserInput struct {
	Login    *string
	Password *string
	Phone    *string
}

func (s *AdminService) CreateOrganization(ctx context.Context, actorID uuid.UUID, input CreateOrganizationInput) (*CreateOrganizationResult, error) {
	if err := validateCreateOrganizationInput(input); err != nil {
		return nil, err
	}

	actor, err := s.users.FindByID(ctx, actorID.String())
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrUserNotFound
	}
	if err != nil {
		return nil, err
	}

	targetType, adminRole, err := resolveOrganizationCreation(actor.Role)
	if err != nil {
		return nil, err
	}

	if actor.OrganizationID == uuid.Nil {
		return nil, ErrHierarchyViolation
	}

	if err := s.ensureParentOrganization(ctx, actor.OrganizationID, actor.Role); err != nil {
		return nil, err
	}

	org := &model.Organization{
		ParentOrgID:  &actor.OrganizationID,
		Type:         targetType,
		Name:         input.Name,
		BIN:          input.BIN,
		HeadFullName: input.HeadFullName,
		Address:      input.Address,
		Phone:        input.Phone,
		IsActive:     true,
	}

	if err := s.organizations.Create(ctx, org); err != nil {
		return nil, err
	}

	adminUser, err := s.buildUserFromAdminInput(ctx, org.ID, adminRole, input.Admin)
	if err != nil {
		return nil, err
	}

	if err := s.users.Create(ctx, adminUser); err != nil {
		return nil, err
	}

	createdUser, err := s.users.FindByID(ctx, adminUser.ID.String())
	if err != nil {
		return nil, err
	}

	return &CreateOrganizationResult{
		Organization: toOrganizationInfo(org),
		Admin:        toUserInfo(createdUser),
	}, nil
}

func (s *AdminService) CreateUser(ctx context.Context, actorID uuid.UUID, input CreateUserInput) (*UserInfo, error) {
	if err := validateCreateUserInput(input); err != nil {
		return nil, err
	}

	actor, err := s.users.FindByID(ctx, actorID.String())
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrUserNotFound
	}
	if err != nil {
		return nil, err
	}

	targetRole, err := resolveUserCreation(actor.Role)
	if err != nil {
		return nil, err
	}

	if actor.OrganizationID == uuid.Nil {
		return nil, ErrHierarchyViolation
	}

	if err := s.ensureParentOrganization(ctx, actor.OrganizationID, actor.Role); err != nil {
		return nil, err
	}

	user := &model.User{
		OrganizationID: actor.OrganizationID,
		Role:           targetRole,
		IsActive:       true,
	}

	if input.Login != nil {
		login := strings.TrimSpace(*input.Login)
		if login != "" {
			exists, err := s.users.ExistsByLogin(ctx, login)
			if err != nil {
				return nil, err
			}
			if exists {
				return nil, ErrConflict
			}
			user.Login = &login
		}
	}

	if input.Phone != nil {
		phone := strings.TrimSpace(*input.Phone)
		if phone != "" {
			exists, err := s.users.ExistsByPhone(ctx, phone)
			if err != nil {
				return nil, err
			}
			if exists {
				return nil, ErrConflict
			}
			user.Phone = &phone
		}
	}

	if input.Password != nil {
		password := strings.TrimSpace(*input.Password)
		if password != "" {
			hash, err := s.password.Hash(password)
			if err != nil {
				return nil, err
			}
			user.PasswordHash = &hash
		}
	}

	if err := s.users.Create(ctx, user); err != nil {
		return nil, err
	}

	createdUser, err := s.users.FindByID(ctx, user.ID.String())
	if err != nil {
		return nil, err
	}

	info := toUserInfo(createdUser)
	return &info, nil
}

func (s *AdminService) buildUserFromAdminInput(ctx context.Context, orgID uuid.UUID, role model.UserRole, input CreateOrganizationAdminInput) (*model.User, error) {
	login, password, phone := "", "", ""

	if input.Login != nil {
		login = strings.TrimSpace(*input.Login)
	}
	if input.Password != nil {
		password = strings.TrimSpace(*input.Password)
	}
	if input.Phone != nil {
		phone = strings.TrimSpace(*input.Phone)
	}

	if login == "" && phone == "" {
		return nil, ErrInvalidInput
	}

	if login != "" {
		exists, err := s.users.ExistsByLogin(ctx, login)
		if err != nil {
			return nil, err
		}
		if exists {
			return nil, ErrConflict
		}
	}

	if phone != "" {
		exists, err := s.users.ExistsByPhone(ctx, phone)
		if err != nil {
			return nil, err
		}
		if exists {
			return nil, ErrConflict
		}
	}

	var passwordHash *string
	if login != "" {
		if password == "" {
			return nil, ErrInvalidInput
		}
		hash, err := s.password.Hash(password)
		if err != nil {
			return nil, err
		}
		passwordHash = &hash
	}

	user := &model.User{
		OrganizationID: orgID,
		Role:           role,
		IsActive:       true,
		PasswordHash:   passwordHash,
	}

	if login != "" {
		user.Login = &login
	}
	if phone != "" {
		user.Phone = &phone
	}

	return user, nil
}

func (s *AdminService) ensureParentOrganization(ctx context.Context, orgID uuid.UUID, role model.UserRole) error {
	org, err := s.organizations.FindByID(ctx, orgID.String())
	if err != nil {
		return err
	}

	switch role {
	case model.UserRoleAkimatAdmin:
		if org.Type != model.OrganizationTypeAkimat {
			return ErrHierarchyViolation
		}
	case model.UserRoleTooAdmin:
		if org.Type != model.OrganizationTypeToo {
			return ErrHierarchyViolation
		}
	case model.UserRoleContractorAdmin, model.UserRoleDriver:
		if org.Type != model.OrganizationTypeContractor {
			return ErrHierarchyViolation
		}
	}
	return nil
}

func toOrganizationInfo(org *model.Organization) OrganizationInfo {
	return OrganizationInfo{
		ID:           org.ID,
		ParentOrgID:  org.ParentOrgID,
		Type:         org.Type,
		Name:         org.Name,
		BIN:          org.BIN,
		HeadFullName: org.HeadFullName,
		Address:      org.Address,
		Phone:        org.Phone,
		IsActive:     org.IsActive,
	}
}

func validateCreateOrganizationInput(input CreateOrganizationInput) error {
	if strings.TrimSpace(input.Name) == "" {
		return ErrInvalidInput
	}
	if err := validateCreateOrganizationAdminInput(input.Admin); err != nil {
		return err
	}
	return nil
}

func validateCreateOrganizationAdminInput(input CreateOrganizationAdminInput) error {
	login := ""
	if input.Login != nil {
		login = strings.TrimSpace(*input.Login)
	}
	password := ""
	if input.Password != nil {
		password = strings.TrimSpace(*input.Password)
	}
	phone := ""
	if input.Phone != nil {
		phone = strings.TrimSpace(*input.Phone)
	}

	if login == "" && phone == "" {
		return ErrInvalidInput
	}
	if login != "" && password == "" {
		return ErrInvalidInput
	}
	return nil
}

func validateCreateUserInput(input CreateUserInput) error {
	login := ""
	if input.Login != nil {
		login = strings.TrimSpace(*input.Login)
	}
	password := ""
	if input.Password != nil {
		password = strings.TrimSpace(*input.Password)
	}
	phone := ""
	if input.Phone != nil {
		phone = strings.TrimSpace(*input.Phone)
	}

	if login == "" && phone == "" {
		return ErrInvalidInput
	}
	if login != "" && password == "" {
		return ErrInvalidInput
	}

	return nil
}

func resolveOrganizationCreation(role model.UserRole) (model.OrganizationType, model.UserRole, error) {
	switch role {
	case model.UserRoleAkimatAdmin:
		return model.OrganizationTypeToo, model.UserRoleTooAdmin, nil
	case model.UserRoleTooAdmin:
		return model.OrganizationTypeContractor, model.UserRoleContractorAdmin, nil
	default:
		return "", "", ErrPermissionDenied
	}
}

func resolveUserCreation(role model.UserRole) (model.UserRole, error) {
	switch role {
	case model.UserRoleContractorAdmin:
		return model.UserRoleDriver, nil
	default:
		return "", ErrPermissionDenied
	}
}
