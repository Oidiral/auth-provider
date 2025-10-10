package repository

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/Oidiral/auth-provider/internal/domain"
	"github.com/jmoiron/sqlx"
)

type Session struct {
	db *sqlx.DB
}

func NewSession(db *sqlx.DB) *Session {
	return &Session{db: db}
}

func (s *Session) Create(ctx context.Context, refreshToken string, userId string, expiresAt time.Time) error {
	_, err := s.db.ExecContext(ctx, "INSERT INTO sessions (refresh_token, user_id, expires_at) VALUES ($1, $2, $3)", refreshToken, userId, expiresAt)
	return err
}

func (s *Session) Delete(ctx context.Context, refreshToken string) error {
	result, err := s.db.ExecContext(ctx, "DELETE FROM sessions WHERE refresh_token = $1", refreshToken)
	if err != nil {
		return err
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return domain.ErrSessionNotFound
	}
	return nil
}

func (s *Session) DeleteAllByUserId(ctx context.Context, userID string) error {
	result, err := s.db.ExecContext(ctx, "DELETE FROM sessions WHERE user_id = $1", userID)
	if err != nil {
		return err
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return domain.ErrSessionNotFound
	}
	return nil
}

func (s *Session) DeleteExpired(ctx context.Context) error {
	_, err := s.db.ExecContext(ctx, "DELETE FROM sessions WHERE expires_at < NOW()")
	return err
}

func (s *Session) Exists(ctx context.Context, refreshToken string) (bool, error) {
	var exists bool
	err := s.db.GetContext(ctx, &exists, "SELECT EXISTS(SELECT 1 FROM sessions WHERE refresh_token = $1)", refreshToken)
	return exists, err
}

func (s *Session) Get(ctx context.Context, refreshToken string) (domain.Session, error) {
	var session domain.Session
	err := s.db.GetContext(ctx, &session, "SELECT * FROM sessions WHERE refresh_token = $1", refreshToken)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return session, domain.ErrSessionNotFound
		}
		return session, err
	}
	return session, nil
}

func (s *Session) GetByUserId(ctx context.Context, userID string) ([]domain.Session, error) {
	var sessions []domain.Session
	err := s.db.SelectContext(ctx, &sessions, "SELECT * FROM sessions WHERE user_id = $1", userID)
	if err != nil {
		return sessions, err
	}
	return sessions, nil
}
