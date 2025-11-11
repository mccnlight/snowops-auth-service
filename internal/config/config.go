package config

import (
	"time"

	"github.com/spf13/viper"

	"github.com/nurpe/snowops-auth/internal/model"
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

	setDefaults(v)

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

	return cfg, nil
}

func setDefaults(v *viper.Viper) {
	v.SetDefault("APP_ENV", "development")
	v.SetDefault("HTTP_HOST", "0.0.0.0")
	v.SetDefault("HTTP_PORT", 8080)

	v.SetDefault("DB_DSN", "postgres://postgres:postgres@localhost:5431/auth_db?sslmode=disable")
	v.SetDefault("DB_MAX_OPEN_CONNS", 25)
	v.SetDefault("DB_MAX_IDLE_CONNS", 10)
	v.SetDefault("DB_CONN_MAX_LIFETIME", time.Hour)

	v.SetDefault("JWT_ACCESS_SECRET", "supersecret")
	v.SetDefault("JWT_ACCESS_TTL", time.Minute*15)
	v.SetDefault("JWT_REFRESH_TTL", time.Hour*24*30)

	v.SetDefault("SMS_CODE_TTL", time.Minute*5)
	v.SetDefault("SMS_CODE_LENGTH", 6)
	v.SetDefault("SMS_DAILY_LIMIT", 10)

	v.SetDefault("ADMIN_SEED_ENABLED", true)
	v.SetDefault("ADMIN_LOGIN", "admin")
	v.SetDefault("ADMIN_PASSWORD", "admin123")
	v.SetDefault("ADMIN_PHONE", "")
	v.SetDefault("ADMIN_ORG_NAME", "Default Akimat")
	v.SetDefault("ADMIN_ORG_BIN", "")

	v.SetDefault("REG_DEFAULT_ROLE", string(model.UserRoleDriver))
	v.SetDefault("REG_DEFAULT_ORG_NAME", "Default Contractor")
	v.SetDefault("REG_DEFAULT_ORG_BIN", "")
}
