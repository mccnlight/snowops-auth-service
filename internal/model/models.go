package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type OrganizationType string

const (
	OrganizationTypeAkimat     OrganizationType = "AKIMAT"
	OrganizationTypeKgu        OrganizationType = "KGU"
	OrganizationTypeContractor OrganizationType = "CONTRACTOR"
)

type UserRole string

const (
	UserRoleAkimatAdmin     UserRole = "AKIMAT_ADMIN"
	UserRoleKguAdmin        UserRole = "KGU_ADMIN"
	UserRoleContractorAdmin UserRole = "CONTRACTOR_ADMIN"
	UserRoleDriver          UserRole = "DRIVER"
)

type Organization struct {
	ID           uuid.UUID        `gorm:"type:uuid;default:uuid_generate_v4();primaryKey"`
	ParentOrgID  *uuid.UUID       `gorm:"type:uuid"`
	Type         OrganizationType `gorm:"type:varchar(20);not null"`
	Name         string           `gorm:"type:varchar(255);not null"`
	BIN          string           `gorm:"type:varchar(12)"`
	HeadFullName string           `gorm:"type:varchar(255)"`
	Address      string           `gorm:"type:varchar(255)"`
	Phone        string           `gorm:"type:varchar(32)"`
	IsActive     bool             `gorm:"default:true"`
	CreatedAt    time.Time        `gorm:"autoCreateTime"`
}

type User struct {
	ID             uuid.UUID  `gorm:"type:uuid;default:uuid_generate_v4();primaryKey"`
	OrganizationID uuid.UUID  `gorm:"type:uuid;not null"`
	Role           UserRole   `gorm:"type:varchar(30);not null"`
	Login          *string    `gorm:"type:varchar(100)"`
	PasswordHash   *string    `gorm:"type:varchar(255)"`
	Phone          *string    `gorm:"type:varchar(32)"`
	DriverID       *uuid.UUID `gorm:"type:uuid"`
	IsActive       bool       `gorm:"default:true"`
	CreatedAt      time.Time  `gorm:"autoCreateTime"`

	Organization Organization `gorm:"foreignKey:OrganizationID"`
}

type SmsCode struct {
	ID        uuid.UUID `gorm:"type:uuid;default:uuid_generate_v4();primaryKey"`
	Phone     string    `gorm:"type:varchar(32);index;not null"`
	Code      string    `gorm:"type:varchar(10);not null"`
	ExpiresAt time.Time `gorm:"not null"`
	IsUsed    bool      `gorm:"default:false"`
	CreatedAt time.Time `gorm:"autoCreateTime"`
}

type UserSession struct {
	ID               uuid.UUID `gorm:"type:uuid;default:uuid_generate_v4();primaryKey"`
	UserID           uuid.UUID `gorm:"type:uuid;not null;index"`
	RefreshTokenHash string    `gorm:"type:char(64);not null;uniqueIndex"`
	ExpiresAt        time.Time `gorm:"not null"`
	RevokedAt        *time.Time
	UserAgent        string    `gorm:"type:varchar(255)"`
	ClientIP         string    `gorm:"type:varchar(45)"`
	CreatedAt        time.Time `gorm:"autoCreateTime"`

	User User `gorm:"foreignKey:UserID"`
}

func (u *User) BeforeCreate(_ *gorm.DB) error {
	if u.ID == uuid.Nil {
		u.ID = uuid.New()
	}
	return nil
}

func (o *Organization) BeforeCreate(_ *gorm.DB) error {
	if o.ID == uuid.Nil {
		o.ID = uuid.New()
	}
	return nil
}

func (s *SmsCode) BeforeCreate(_ *gorm.DB) error {
	if s.ID == uuid.Nil {
		s.ID = uuid.New()
	}
	return nil
}

func (s *UserSession) BeforeCreate(_ *gorm.DB) error {
	if s.ID == uuid.Nil {
		s.ID = uuid.New()
	}
	return nil
}
