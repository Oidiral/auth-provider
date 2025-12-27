package service

import (
	"context"
	"time"

	"github.com/Oidiral/auth-provider/internal/repository"
	"github.com/Oidiral/auth-provider/pkg/auth"
	"github.com/Oidiral/auth-provider/pkg/logger"
	"github.com/Oidiral/auth-provider/pkg/otp"
)

type UserSignUpInput struct {
	Email     string
	Phone     string
	Username  string
	FirstName string
	LastName  string
	Password  string
}

type UserSignInInput struct {
	Email    string
	Password string
}

type Tokens struct {
	AccessToken  string
	RefreshToken string
}

type Users interface {
	SignUp(ctx context.Context, input UserSignUpInput) error
	SignIn(ctx context.Context, input UserSignInInput) (Tokens, error)
	Refresh(ctx context.Context, refreshToken string) (Tokens, error)
	Verify(ctx context.Context, userId string, hash string) error
	OtpRetrySend(ctx context.Context, userId string) error
}

// Services - контейнер всех сервисов приложения
// Используется для удобной передачи в HTTP handlers
type Services struct {
	Users Users
	// Здесь будут добавляться другие сервисы по мере роста приложения:
	// Admin Admin
	// TwoFA TwoFA
}

// Deps - все зависимости, необходимые для создания сервисов
// Собирает все зависимости в одном месте для удобства
type Deps struct {
	// Repositories
	Repos *repository.Repositories

	// External services
	TokenManager auth.TokenManager
	OtpGenerator otp.Generator

	// Infrastructure
	Logger logger.Logger

	// Configuration
	AccessTokenTTL  time.Duration
	RefreshTokenTTL time.Duration
}

// NewServices - создает все сервисы приложения
// Вызывается один раз при инициализации приложения
func NewServices(deps Deps) *Services {
	userService := NewUserService(
		deps.Repos.Users,
		deps.TokenManager,
		deps.AccessTokenTTL,
		deps.RefreshTokenTTL,
		deps.Repos.Sessions,
		deps.Repos.OtpCodes,
		deps.Logger,
	)

	return &Services{
		Users: userService,
	}
}
