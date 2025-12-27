package app

import (
	"context"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Oidiral/auth-provider/internal/config"
	httpdelivery "github.com/Oidiral/auth-provider/internal/delivery/http"
	"github.com/Oidiral/auth-provider/internal/infrastructure/cache"
	"github.com/Oidiral/auth-provider/internal/infrastructure/database/postgres"
	"github.com/Oidiral/auth-provider/internal/repository"
	"github.com/Oidiral/auth-provider/internal/service"
	"github.com/Oidiral/auth-provider/pkg/auth"
	"github.com/Oidiral/auth-provider/pkg/logger"
	"github.com/Oidiral/auth-provider/pkg/otp"
)

func Run() {
	log := logger.NewZerologLogger(os.Stdout)

	cfg, err := config.Init(log)
	if err != nil {
		log.Error("failed to initialize config", err)
		return
	}

	postgresDB, err := postgres.NewClient(cfg.Postgres.Host, cfg.Postgres.Port, cfg.Postgres.User, cfg.Postgres.Password, cfg.Postgres.DBName, log)
	if err != nil {
		log.Error("failed to connect to postgres", err)
		return
	}
	defer func() {
		if err := postgresDB.Close(); err != nil {
			log.Error("failed to close postgres client", err)
		}
	}()

	if err := postgres.RunMigrations(postgres.GetSqlDB(postgresDB), log, postgres.MigrationsConfig{}); err != nil {
		log.Error("failed to apply migrations", err)
		return
	}

	redis, err := cache.NewClient(cfg.Redis, log)
	if err != nil {
		log.Error("failed to connect to redis", err)
		return
	}
	defer func() {
		if err := redis.Close(); err != nil {
			log.Error("failed to close redis client", err)
		}
	}()

	otpGenerator := otp.NewGOTPGenerator()

	tokentManager, err := auth.NewManager(cfg.Auth.JWT.SigningKey, log)
	if err != nil {
		log.Error("failed to create token manager", err)
		return
	}

	repos := repository.NewRepositories(postgresDB, redis, cfg.Auth.ExpiresSessionTTL, cfg.Auth.ExpiresOtpTTL, log)

	services := service.NewServices(service.Deps{
		Repos:           repos,
		Logger:          log,
		TokenManager:    tokentManager,
		OtpGenerator:    otpGenerator,
		AccessTokenTTL:  cfg.Auth.JWT.AccessTokenTTL,
		RefreshTokenTTL: cfg.Auth.JWT.RefreshTokenTTL,
	})
	handler := httpdelivery.NewHandler(services)
	router := handler.Init(log)

	srv := &http.Server{
		Addr:    ":" + cfg.HTTP.Port,
		Handler: router,
	}

	log.Info("starting HTTP server", logger.Field{Key: "port", Value: cfg.HTTP.Port})

	go func() {
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Error("failed to start HTTP server", err)
		}
	}()

	log.Info("application initialized successfully")

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit

	log.Info("shutting down application")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Error("failed to shut down HTTP server", err)
	}

	log.Info("application shut down successfully")
}
