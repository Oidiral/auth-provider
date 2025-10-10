package repository

import (
	"context"
	"time"

	"github.com/Oidiral/auth-provider/internal/domain"
	"github.com/jmoiron/sqlx"
)

type Sessions interface {
	Create(ctx context.Context, refreshToken string, userId string, expiresAt time.Time) error
	Delete(ctx context.Context, refreshToken string) error
	DeleteAllByUserId(ctx context.Context, userID string) error
	DeleteExpired(ctx context.Context) error
	Exists(ctx context.Context, refreshToken string) (bool, error)
	Get(ctx context.Context, refreshToken string) (domain.Session, error)
	GetByUserId(ctx context.Context, userID string) ([]domain.Session, error)
}

type Users interface {
	Create(ctx context.Context, input domain.User) error
	Get(ctx context.Context, id string) (domain.User, error)
	GetByEmail(ctx context.Context, email string) (domain.User, error)
	Update(ctx context.Context, id string, input domain.User) error
	Delete(ctx context.Context, id string) error
}

type Repositories struct {
	Sessions Sessions
	Users    Users
}

func NewRepositories(db *sqlx.DB) *Repositories {
	return &Repositories{
		Sessions: NewSession(db),
	}
}
