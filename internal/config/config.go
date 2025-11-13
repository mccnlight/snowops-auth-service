package config

import (
	"fmt"
	"time"

	"github.com/spf13/viper"
)

type HTTPConfig struct {
	Host string
	Port int
}

type DBConfig struct {
	DSN             string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
}

type JWTConfig struct {
	AccessSecret string
	AccessTTL    time.Duration
	RefreshTTL   time.Duration
}

type SMSConfig struct {
	CodeTTL        time.Duration
	CodeLength     int
	DailySendLimit int
}

type AdminSeedConfig struct {
	Enabled          bool
	Login            string
	Password         string
	Phone            string
	OrganizationName string
	OrganizationBIN  string
}

type RegistrationConfig struct {
	DefaultRole            string
	DefaultOrganization    string
	DefaultOrganizationBIN string
}

type Config struct {
	Environment  string
	HTTP         HTTPConfig
	DB           DBConfig
	JWT          JWTConfig
	SMS          SMSConfig
	Admin        AdminSeedConfig
	Registration RegistrationConfig
}

func Load() (*Config, error) {
	v := viper.New()
	v.SetConfigName("app")
	v.SetConfigType("env")
	v.AddConfigPath(".")
	v.AddConfigPath("./config")
	v.AddConfigPath("./deploy")
	v.AddConfigPath("./internal/config")

	v.AutomaticEnv()

	_ = v.ReadInConfig()

	cfg := &Config{
		Environment: v.GetString("APP_ENV"),
		HTTP: HTTPConfig{
			Host: v.GetString("HTTP_HOST"),
			Port: v.GetInt("HTTP_PORT"),
		},
		DB: DBConfig{
			DSN:             v.GetString("DB_DSN"),
			MaxOpenConns:    v.GetInt("DB_MAX_OPEN_CONNS"),
			MaxIdleConns:    v.GetInt("DB_MAX_IDLE_CONNS"),
			ConnMaxLifetime: v.GetDuration("DB_CONN_MAX_LIFETIME"),
		},
		JWT: JWTConfig{
			AccessSecret: v.GetString("JWT_ACCESS_SECRET"),
			AccessTTL:    v.GetDuration("JWT_ACCESS_TTL"),
			RefreshTTL:   v.GetDuration("JWT_REFRESH_TTL"),
		},
		SMS: SMSConfig{
			CodeTTL:        v.GetDuration("SMS_CODE_TTL"),
			CodeLength:     v.GetInt("SMS_CODE_LENGTH"),
			DailySendLimit: v.GetInt("SMS_DAILY_LIMIT"),
		},
		Admin: AdminSeedConfig{
			Enabled:          v.GetBool("ADMIN_SEED_ENABLED"),
			Login:            v.GetString("ADMIN_LOGIN"),
			Password:         v.GetString("ADMIN_PASSWORD"),
			Phone:            v.GetString("ADMIN_PHONE"),
			OrganizationName: v.GetString("ADMIN_ORG_NAME"),
			OrganizationBIN:  v.GetString("ADMIN_ORG_BIN"),
		},
		Registration: RegistrationConfig{
			DefaultRole:            v.GetString("REG_DEFAULT_ROLE"),
			DefaultOrganization:    v.GetString("REG_DEFAULT_ORG_NAME"),
			DefaultOrganizationBIN: v.GetString("REG_DEFAULT_ORG_BIN"),
		},
	}

	if err := validate(cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}

func validate(cfg *Config) error {
	if cfg.DB.DSN == "" {
		return fmt.Errorf("DB_DSN is required")
	}
	if cfg.JWT.AccessSecret == "" {
		return fmt.Errorf("JWT_ACCESS_SECRET is required")
	}
	if cfg.JWT.AccessTTL == 0 {
		return fmt.Errorf("JWT_ACCESS_TTL is required")
	}
	if cfg.JWT.RefreshTTL == 0 {
		return fmt.Errorf("JWT_REFRESH_TTL is required")
	}
	if cfg.HTTP.Host == "" {
		return fmt.Errorf("HTTP_HOST is required")
	}
	if cfg.HTTP.Port == 0 {
		return fmt.Errorf("HTTP_PORT is required")
	}
	if cfg.SMS.CodeTTL == 0 {
		return fmt.Errorf("SMS_CODE_TTL is required")
	}
	if cfg.SMS.CodeLength == 0 {
		return fmt.Errorf("SMS_CODE_LENGTH is required")
	}
	return nil
}
