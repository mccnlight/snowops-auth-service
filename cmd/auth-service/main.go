package main

import (
	"fmt"
	"os"

	"github.com/nurpe/snowops-auth/internal/adapter/password"
	"github.com/nurpe/snowops-auth/internal/adapter/sms"
	"github.com/nurpe/snowops-auth/internal/adapter/token"
	"github.com/nurpe/snowops-auth/internal/config"
	"github.com/nurpe/snowops-auth/internal/db"
	httphandler "github.com/nurpe/snowops-auth/internal/http"
	"github.com/nurpe/snowops-auth/internal/http/middleware"
	"github.com/nurpe/snowops-auth/internal/logger"
	"github.com/nurpe/snowops-auth/internal/repository"
	"github.com/nurpe/snowops-auth/internal/service"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to load config: %v\n", err)
		os.Exit(1)
	}

	appLogger := logger.New(cfg.Environment)

	database, err := db.New(cfg, appLogger)
	if err != nil {
		appLogger.Fatal().Err(err).Msg("failed to connect database")
	}

	userRepo := repository.NewUserRepository(database)
	smsRepo := repository.NewSmsCodeRepository(database)
	sessionRepo := repository.NewUserSessionRepository(database)

	passwordHasher := password.NewBcryptHasher(0)
	tokenManager := token.NewManager(cfg.JWT.AccessSecret)
	smsSender := sms.NewLoggerSender(appLogger)

	authService := service.NewAuthService(
		userRepo,
		sessionRepo,
		smsRepo,
		passwordHasher,
		smsSender,
		tokenManager,
		cfg,
	)

	handler := httphandler.NewHandler(authService, appLogger)
	authMiddleware := middleware.Auth(tokenManager)
	router := httphandler.NewRouter(handler, authMiddleware, cfg.Environment)

	addr := fmt.Sprintf("%s:%d", cfg.HTTP.Host, cfg.HTTP.Port)
	appLogger.Info().Str("addr", addr).Msg("starting auth service")

	if err := router.Run(addr); err != nil {
		appLogger.Error().Err(err).Msg("failed to start server")
		os.Exit(1)
	}
}
