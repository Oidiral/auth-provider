package repository

import (
	"context"
	"time"

	"github.com/Oidiral/auth-provider/internal/domain"
	"github.com/Oidiral/auth-provider/pkg/logger"
	"github.com/jmoiron/sqlx"
	"github.com/redis/go-redis/v9"
)

type Sessions interface {
	Create(ctx context.Context, refreshToken string, userId string) error
	Delete(ctx context.Context, refreshToken string) error
	DeleteAllByUserId(ctx context.Context, userID string) error
	Exists(ctx context.Context, refreshToken string) (bool, error)
	Get(ctx context.Context, refreshToken string) (domain.Session, error)
	GetByUserId(ctx context.Context, userID string) ([]domain.Session, error)
	Replace(ctx context.Context, oldToken, newToken, userID string) error
}

type OtpCodesRepository interface {
	Create(ctx context.Context, userId string) error
	Delete(ctx context.Context, userId string) error
	Get(ctx context.Context, userId string) (domain.OTP, error)
	Verify(ctx context.Context, userId, code string) error
}

type Users interface {
	Create(ctx context.Context, input domain.User) (string, error)
	Get(ctx context.Context, id string) (domain.User, error)
	GetByEmail(ctx context.Context, email string) (domain.User, error)
	Update(ctx context.Context, id string, input domain.UpdateUserInput) error
	Delete(ctx context.Context, id string) error
}

type Repositories struct {
	Sessions Sessions
	Users    Users
	OtpCodes OtpCodesRepository
}

func NewRepositories(db *sqlx.DB, redisClient *redis.Client, ExpiresSessionTTL time.Duration, ExpiresOtpTTl time.Duration, log logger.Logger) *Repositories {
	return &Repositories{
		Sessions: NewSession(redisClient, ExpiresSessionTTL, log),
		OtpCodes: NewOtpManager(redisClient, log, ExpiresOtpTTl),
		Users:    NewUserRepo(db, log),
	}
}
