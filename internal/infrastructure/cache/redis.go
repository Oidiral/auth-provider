package cache

import (
	"context"
	"fmt"
	"time"

	"github.com/Oidiral/auth-provider/internal/config"
	"github.com/Oidiral/auth-provider/pkg/logger"
	"github.com/redis/go-redis/v9"
)

func NewClient(cfg config.RedisConfig, log logger.Logger) (*redis.Client, error) {
	log.Info("connecting to redis", logger.Field{Key: "host", Value: cfg.Host}, logger.Field{Key: "port", Value: cfg.Port})

	client := redis.NewClient(&redis.Options{
		Addr:         fmt.Sprintf("%s:%s", cfg.Host, cfg.Port),
		Password:     cfg.Password,
		DB:           0,
		DialTimeout:  5 * time.Second,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
		PoolSize:     10,
		MinIdleConns: 5,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		log.Error("failed to connect to redis", err)
		return nil, fmt.Errorf("failed to connect to redis: %w", err)
	}

	log.Info("successfully connected to redis")
	return client, nil
}
