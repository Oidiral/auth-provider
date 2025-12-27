package repository

import (
	"context"
	"time"

	"github.com/Oidiral/auth-provider/internal/domain"
	"github.com/Oidiral/auth-provider/pkg/logger"
	"github.com/jmoiron/sqlx"
	"github.com/redis/go-redis/v9"
)

// Sessions определяет интерфейс для работы с сессиями пользователей
type Sessions interface {
	// Create создает новую сессию с указанным refresh токеном для пользователя
	Create(ctx context.Context, refreshToken string, userId string) error
	// Delete удаляет сессию по refresh токену
	Delete(ctx context.Context, refreshToken string) error
	// DeleteAllByUserId удаляет все сессии пользователя
	DeleteAllByUserId(ctx context.Context, userID string) error
	// Exists проверяет существование сессии по refresh токену
	Exists(ctx context.Context, refreshToken string) (bool, error)
	// Get получает сессию по refresh токену
	Get(ctx context.Context, refreshToken string) (domain.Session, error)
	// GetByUserId получает все активные сессии пользователя
	GetByUserId(ctx context.Context, userID string) ([]domain.Session, error)
	// Replace заменяет старый refresh токен на новый для указанного пользователя
	Replace(ctx context.Context, oldToken, newToken, userID string) error
}

// OtpCodesRepository определяет интерфейс для работы с OTP кодами
type OtpCodesRepository interface {
	// Create создает новый OTP код для пользователя
	Create(ctx context.Context, userId string) error
	// Delete удаляет OTP код пользователя
	Delete(ctx context.Context, userId string) error
	// Get получает OTP код пользователя
	Get(ctx context.Context, userId string) (domain.OTP, error)
	// Verify проверяет и подтверждает OTP код пользователя
	Verify(ctx context.Context, userId, code string) error
}

// Users определяет интерфейс для работы с пользователями
type Users interface {
	// Create создает нового пользователя и возвращает его ID
	Create(ctx context.Context, input domain.User) (string, error)
	// Get получает пользователя по ID
	Get(ctx context.Context, id string) (domain.User, error)
	// GetByEmail получает пользователя по email адресу
	GetByEmail(ctx context.Context, email string) (domain.User, error)
	// Update обновляет данные пользователя
	Update(ctx context.Context, id string, input domain.UpdateUserInput) error
	// Delete удаляет пользователя по ID
	Delete(ctx context.Context, id string) error
}

// Repositories объединяет все репозитории приложения
type Repositories struct {
	Sessions Sessions
	Users    Users
	OtpCodes OtpCodesRepository
}

// NewRepositories создает новый экземпляр Repositories с инициализированными репозиториями
func NewRepositories(db *sqlx.DB, redisClient *redis.Client, ExpiresSessionTTL time.Duration, ExpiresOtpTTl time.Duration, log logger.Logger) *Repositories {
	return &Repositories{
		Sessions: NewSession(redisClient, ExpiresSessionTTL, log),
		OtpCodes: NewOtpManager(redisClient, log, ExpiresOtpTTl),
		Users:    NewUserRepo(db, log),
	}
}
