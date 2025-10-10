package app

import (
	"github.com/Oidiral/auth-provider/internal/config"
	"github.com/Oidiral/auth-provider/internal/infrastructure/cache"
	"github.com/Oidiral/auth-provider/internal/infrastructure/database/postgres"
	"github.com/Oidiral/auth-provider/pkg/auth"
	"github.com/Oidiral/auth-provider/pkg/logger"
)

func Run() {
	cfg, err := config.Init()
	if err != nil {
		logger.Error(err)
		return
	}

	// DependencyInjection
	postgres, err := postgres.NewClient(cfg.Postgres.Host, cfg.Postgres.Port, cfg.Postgres.User, cfg.Postgres.Password, cfg.Postgres.DBName)
	if err != nil {
		logger.Errorf("failed to connect to postgres: %v", err)
		return
	}

	redis, err := cache.NewClient(cfg.Redis)
	if err != nil {
		logger.Errorf("failed to connect to redis: %v", err)
		return
	}

	// TODO: Сделать подключение к email микросервису

	tokentManager, err := auth.NewManager(cfg.Auth.JWT.SigningKey)
	if err != nil {
		logger.Errorf("failed to create token manager: %v", err)
		return
	}

	repos :=
}
