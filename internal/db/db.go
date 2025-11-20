package db

import (
	"context"
	"errors"
	"time"

	"github.com/rs/zerolog"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"

	"github.com/nurpe/snowops-auth/internal/adapter/password"
	"github.com/nurpe/snowops-auth/internal/config"
	"github.com/nurpe/snowops-auth/internal/model"
)

func New(cfg *config.Config, log zerolog.Logger) (*gorm.DB, error) {
	dbCfg := cfg.DB
	gormLog := gormlogger.New(
		zerologWriter{logger: log},
		gormlogger.Config{
			SlowThreshold:             time.Second,
			Colorful:                  false,
			IgnoreRecordNotFoundError: true,
			LogLevel:                  selectLogLevel(cfg.Environment),
		},
	)

	db, err := gorm.Open(postgres.Open(dbCfg.DSN), &gorm.Config{
		Logger:      gormLog,
		PrepareStmt: false, // Disable prepared statements to avoid "cached plan must not change result type" errors
	})
	if err != nil {
		return nil, err
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}

	if dbCfg.MaxOpenConns > 0 {
		sqlDB.SetMaxOpenConns(dbCfg.MaxOpenConns)
	}
	if dbCfg.MaxIdleConns > 0 {
		sqlDB.SetMaxIdleConns(dbCfg.MaxIdleConns)
	}
	if dbCfg.ConnMaxLifetime > 0 {
		sqlDB.SetConnMaxLifetime(dbCfg.ConnMaxLifetime)
	}

	if err := db.Exec(`CREATE EXTENSION IF NOT EXISTS "uuid-ossp"`).Error; err != nil {
		return nil, err
	}
	if err := db.Exec(`CREATE EXTENSION IF NOT EXISTS "pgcrypto"`).Error; err != nil {
		return nil, err
	}

	if err := db.AutoMigrate(
		&model.Organization{},
		&model.User{},
		&model.SmsCode{},
		&model.UserSession{},
	); err != nil {
		return nil, err
	}

	if cfg.Admin.Enabled {
		if err := seedAdmin(db, cfg.Admin, log); err != nil {
			return nil, err
		}
	}

	return db, nil
}

func HealthCheck(ctx context.Context, db *gorm.DB) error {
	return db.WithContext(ctx).Exec("SELECT 1").Error
}

func selectLogLevel(env string) gormlogger.LogLevel {
	if env == "development" {
		return gormlogger.Info
	}
	return gormlogger.Warn
}

type zerologWriter struct {
	logger zerolog.Logger
}

func (w zerologWriter) Printf(msg string, args ...interface{}) {
	w.logger.Info().Msgf(msg, args...)
}

func seedAdmin(db *gorm.DB, adminCfg config.AdminSeedConfig, log zerolog.Logger) error {
	if adminCfg.Login == "" || adminCfg.Password == "" {
		log.Warn().Msg("admin seed skipped: login or password not set")
		return nil
	}

	var existing model.User
	if err := db.Where("role = ? AND login = ?", model.UserRoleAkimatAdmin, adminCfg.Login).First(&existing).Error; err == nil {
		return nil
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}

	hasher := password.NewBcryptHasher(0)
	hash, err := hasher.Hash(adminCfg.Password)
	if err != nil {
		return err
	}

	orgName := adminCfg.OrganizationName
	if orgName == "" {
		orgName = "Default Akimat"
	}

	return db.Transaction(func(tx *gorm.DB) error {
		var org model.Organization
		if err := tx.Where("name = ?", orgName).First(&org).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				org = model.Organization{
					Name:     orgName,
					Type:     model.OrganizationTypeAkimat,
					BIN:      adminCfg.OrganizationBIN,
					IsActive: true,
				}
				if err := tx.Create(&org).Error; err != nil {
					return err
				}
			} else {
				return err
			}
		}

		login := adminCfg.Login
		passwordHash := hash
		user := model.User{
			OrganizationID: org.ID,
			Role:           model.UserRoleAkimatAdmin,
			Login:          &login,
			PasswordHash:   &passwordHash,
			IsActive:       true,
		}

		if adminCfg.Phone != "" {
			phone := adminCfg.Phone
			user.Phone = &phone
		}

		if err := tx.Create(&user).Error; err != nil {
			return err
		}

		log.Info().
			Str("login", adminCfg.Login).
			Str("organization", orgName).
			Msg("seeded default admin user")

		return nil
	})
}
