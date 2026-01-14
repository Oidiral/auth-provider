package repository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/Oidiral/auth-provider/internal/domain"
	"github.com/Oidiral/auth-provider/pkg/logger"
	"github.com/redis/go-redis/v9"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
)

type Session struct {
	redis      *redis.Client
	ExpiresTTL time.Duration
	logger     logger.Logger
}

func NewSession(redis *redis.Client, expiresTTL time.Duration, log logger.Logger) *Session {
	if expiresTTL <= 0 {
		panic("session TTL must be positive")
	}
	return &Session{
		redis:      redis,
		ExpiresTTL: expiresTTL,
		logger:     log,
	}
}

func (s *Session) sessionKey(refreshToken string) string {
	return fmt.Sprintf("session:%s", refreshToken)
}

func (s *Session) userSessionsKey(userID string) string {
	return fmt.Sprintf("user_sessions:%s", userID)
}

func (s *Session) Create(ctx context.Context, refreshToken string, userID string) error {
	ctx, span := otel.Tracer("session-repo").Start(ctx, "Session.Create")
	defer span.End()

	span.SetAttributes(
		attribute.String("db.system", "redis"),
		attribute.String("user.id", userID),
	)
	log := s.logger.WithContext(ctx)

	session := domain.Session{
		UserID:       userID,
		RefreshToken: refreshToken,
	}

	sessionData, err := json.Marshal(session)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to marshal session")
		log.Error("failed to marshal session", err, logger.Field{Key: "user_id", Value: userID})
		return fmt.Errorf("failed to marshal session: %w", err)
	}

	pipe := s.redis.TxPipeline()

	pipe.Set(ctx, s.sessionKey(refreshToken), sessionData, s.ExpiresTTL)

	pipe.SAdd(ctx, s.userSessionsKey(userID), refreshToken)

	pipe.Expire(ctx, s.userSessionsKey(userID), s.ExpiresTTL+24*time.Hour)

	_, err = pipe.Exec(ctx)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to create session")
		log.Error("failed to create session in redis", err, logger.Field{Key: "user_id", Value: userID})
		return fmt.Errorf("failed to create session: %w", err)
	}

	return nil
}

func (s *Session) Delete(ctx context.Context, refreshToken string) error {
	ctx, span := otel.Tracer("session-repo").Start(ctx, "Session.Delete")
	defer span.End()

	span.SetAttributes(attribute.String("db.system", "redis"))
	log := s.logger.WithContext(ctx)

	session, err := s.Get(ctx, refreshToken)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to get session")
		log.Error("failed to get session for deletion", err)
		return err
	}
	span.SetAttributes(attribute.String("user.id", session.UserID))

	pipe := s.redis.TxPipeline()

	pipe.Del(ctx, s.sessionKey(refreshToken))

	pipe.SRem(ctx, s.userSessionsKey(session.UserID), refreshToken)

	_, err = pipe.Exec(ctx)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to delete session")
		log.Error("failed to delete session from redis", err, logger.Field{Key: "user_id", Value: session.UserID})
		return fmt.Errorf("failed to delete session: %w", err)
	}

	return nil
}

func (s *Session) DeleteAllByUserId(ctx context.Context, userID string) error {
	log := s.logger.WithContext(ctx)

	refreshTokens, err := s.redis.SMembers(ctx, s.userSessionsKey(userID)).Result()
	if err != nil {
		log.Error("failed to get user sessions", err, logger.Field{Key: "user_id", Value: userID})
		return fmt.Errorf("failed to get user sessions: %w", err)
	}

	if len(refreshTokens) == 0 {
		return domain.ErrSessionNotFound
	}

	pipe := s.redis.TxPipeline()

	for _, token := range refreshTokens {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
		pipe.Del(ctx, s.sessionKey(token))
	}

	pipe.Del(ctx, s.userSessionsKey(userID))

	_, err = pipe.Exec(ctx)
	if err != nil {
		log.Error("failed to delete user sessions from redis", err, logger.Field{Key: "user_id", Value: userID})
		return fmt.Errorf("failed to delete user sessions: %w", err)
	}

	return nil
}

func (s *Session) Exists(ctx context.Context, refreshToken string) (bool, error) {
	exists, err := s.redis.Exists(ctx, s.sessionKey(refreshToken)).Result()
	if err != nil {
		return false, fmt.Errorf("failed to check session existence: %w", err)
	}
	return exists > 0, nil
}

func (s *Session) Get(ctx context.Context, refreshToken string) (domain.Session, error) {
	ctx, span := otel.Tracer("session-repo").Start(ctx, "Session.Get")
	defer span.End()

	span.SetAttributes(attribute.String("db.system", "redis"))

	var session domain.Session

	data, err := s.redis.Get(ctx, s.sessionKey(refreshToken)).Result()
	if err != nil {
		span.RecordError(err)
		if errors.Is(err, redis.Nil) {
			span.SetStatus(codes.Error, "session not found")
			return session, domain.ErrSessionNotFound
		}
		span.SetStatus(codes.Error, "failed to get session")
		return session, fmt.Errorf("failed to get session: %w", err)
	}

	if err := json.Unmarshal([]byte(data), &session); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to unmarshal session")
		return session, fmt.Errorf("failed to unmarshal session: %w", err)
	}

	span.SetAttributes(attribute.String("user.id", session.UserID))
	return session, nil
}

func (s *Session) GetByUserId(ctx context.Context, userID string) ([]domain.Session, error) {
	refreshTokens, err := s.redis.SMembers(ctx, s.userSessionsKey(userID)).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get user sessions: %w", err)
	}

	sessions := make([]domain.Session, 0, len(refreshTokens))

	for _, token := range refreshTokens {
		select {
		case <-ctx.Done():
			return sessions, ctx.Err()
		default:
		}
		session, err := s.Get(ctx, token)
		if err != nil {
			if errors.Is(err, domain.ErrSessionNotFound) || errors.Is(err, domain.ErrSessionExpired) {
				continue
			}
			return nil, err
		}
		sessions = append(sessions, session)
	}

	return sessions, nil
}

func (s *Session) Replace(ctx context.Context, oldToken, newToken, userID string) error {
	ctx, span := otel.Tracer("session-repo").Start(ctx, "Session.Replace")
	defer span.End()

	span.SetAttributes(
		attribute.String("db.system", "redis"),
		attribute.String("user.id", userID),
	)
	log := s.logger.WithContext(ctx)

	script := redis.NewScript(`
		if redis.call('EXISTS', KEYS[1]) == 0 then
			return redis.error_reply('session not found')
		end

		redis.call('DEL', KEYS[1])

		redis.call('SET', KEYS[2], ARGV[1], 'EX', ARGV[2])

		redis.call('SREM', KEYS[3], ARGV[4])

		redis.call('SADD', KEYS[3], ARGV[5])

		redis.call('EXPIRE', KEYS[3], ARGV[3])

		return 1
	`)

	newSession := domain.Session{
		UserID:       userID,
		RefreshToken: newToken,
	}

	sessionData, err := json.Marshal(newSession)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to marshal session")
		log.Error("failed to marshal session", err, logger.Field{Key: "user_id", Value: userID})
		return fmt.Errorf("failed to marshal session: %w", err)
	}

	sessionTTL := int(s.ExpiresTTL.Seconds())
	userSessionsTTL := int((s.ExpiresTTL + 24*time.Hour).Seconds())

	err = script.Run(ctx, s.redis,
		[]string{
			s.sessionKey(oldToken),
			s.sessionKey(newToken),
			s.userSessionsKey(userID),
		},
		sessionData,
		sessionTTL,
		userSessionsTTL,
		oldToken,
		newToken,
	).Err()

	if err != nil {
		span.RecordError(err)
		if err.Error() == "session not found" {
			span.SetStatus(codes.Error, "session not found")
			return domain.ErrSessionNotFound
		}
		span.SetStatus(codes.Error, "failed to replace session")
		log.Error("failed to replace session", err, logger.Field{Key: "user_id", Value: userID})
		return fmt.Errorf("failed to replace session: %w", err)
	}

	return nil
}
