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
	// Create создает новый OTP код для пользователя и возвращает его
	Create(ctx context.Context, userId string) (string, error)
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

// Roles определяет интерфейс для работы с ролями пользователей
type Roles interface {
	// GetById получает роль по ID
	GetById(ctx context.Context, id int) (domain.Role, error)
	// GetByName получает роль по имени
	GetByName(ctx context.Context, name string) (domain.Role, error)
	// GetUserRoles получает все роли пользователя
	GetUserRoles(ctx context.Context, userId string) ([]domain.Role, error)
	// AssignToUser назначает роль пользователю
	AssignToUser(ctx context.Context, userId string, roleId int) error
	// RemoveFromUser удаляет роль у пользователя
	RemoveFromUser(ctx context.Context, userId string, roleId int) error
	// UserHasRole проверяет наличие роли у пользователя
	UserHasRole(ctx context.Context, userId string, roleId int) (bool, error)
}

// UnitOfWork определяет интерфейс для атомарных операций в рамках транзакции
type UnitOfWork interface {
	// Users возвращает репозиторий пользователей в рамках транзакции
	Users() Users
	// Roles возвращает репозиторий ролей в рамках транзакции
	Roles() Roles
	// Commit фиксирует транзакцию
	Commit() error
	// Rollback откатывает транзакцию
	Rollback() error
}

// UoWFactory определяет интерфейс для создания Unit of Work
type UoWFactory interface {
	// Begin начинает новую транзакцию и возвращает UnitOfWork
	Begin(ctx context.Context) (UnitOfWork, error)
}

// Repositories объединяет все репозитории приложения
type Repositories struct {
	Sessions   Sessions
	Users      Users
	Roles      Roles
	OtpCodes   OtpCodesRepository
	UoWFactory UoWFactory
}

// NewRepositories создает новый экземпляр Repositories с инициализированными репозиториями
func NewRepositories(db *sqlx.DB, redisClient *redis.Client, ExpiresSessionTTL time.Duration, ExpiresOtpTTl time.Duration, log logger.Logger) *Repositories {
	return &Repositories{
		Sessions:   NewSession(redisClient, ExpiresSessionTTL, log),
		OtpCodes:   NewOtpRepository(redisClient, log, ExpiresOtpTTl),
		Users:      NewUserRepo(db, log),
		Roles:      NewRole(db, log),
		UoWFactory: NewUoWFactory(db, log),
	}
}
