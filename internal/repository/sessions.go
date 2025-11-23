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
	log := s.logger.WithContext(ctx)
	log.Info("creating session", logger.Field{Key: "user_id", Value: userID})

	session := domain.Session{
		UserID:       userID,
		RefreshToken: refreshToken,
	}

	sessionData, err := json.Marshal(session)
	if err != nil {
		log.Error("failed to marshal session", err, logger.Field{Key: "user_id", Value: userID})
		return fmt.Errorf("failed to marshal session: %w", err)
	}

	pipe := s.redis.TxPipeline()

	pipe.Set(ctx, s.sessionKey(refreshToken), sessionData, s.ExpiresTTL)

	pipe.SAdd(ctx, s.userSessionsKey(userID), refreshToken)

	pipe.Expire(ctx, s.userSessionsKey(userID), s.ExpiresTTL+24*time.Hour)

	_, err = pipe.Exec(ctx)
	if err != nil {
		log.Error("failed to create session in redis", err, logger.Field{Key: "user_id", Value: userID})
		return fmt.Errorf("failed to create session: %w", err)
	}

	log.Info("session created successfully", logger.Field{Key: "user_id", Value: userID})
	return nil
}

func (s *Session) Delete(ctx context.Context, refreshToken string) error {
	log := s.logger.WithContext(ctx)
	log.Info("deleting session")

	session, err := s.Get(ctx, refreshToken)
	if err != nil {
		log.Error("failed to get session for deletion", err)
		return err
	}

	pipe := s.redis.TxPipeline()

	pipe.Del(ctx, s.sessionKey(refreshToken))

	pipe.SRem(ctx, s.userSessionsKey(session.UserID), refreshToken)

	_, err = pipe.Exec(ctx)
	if err != nil {
		log.Error("failed to delete session from redis", err, logger.Field{Key: "user_id", Value: session.UserID})
		return fmt.Errorf("failed to delete session: %w", err)
	}

	log.Info("session deleted successfully", logger.Field{Key: "user_id", Value: session.UserID})
	return nil
}

func (s *Session) DeleteAllByUserId(ctx context.Context, userID string) error {
	log := s.logger.WithContext(ctx)
	log.Info("deleting all sessions", logger.Field{Key: "user_id", Value: userID})

	refreshTokens, err := s.redis.SMembers(ctx, s.userSessionsKey(userID)).Result()
	if err != nil {
		log.Error("failed to get user sessions", err, logger.Field{Key: "user_id", Value: userID})
		return fmt.Errorf("failed to get user sessions: %w", err)
	}

	if len(refreshTokens) == 0 {
		log.Warn("no sessions found", logger.Field{Key: "user_id", Value: userID})
		return domain.ErrSessionNotFound
	}

	pipe := s.redis.TxPipeline()

	for _, token := range refreshTokens {
		pipe.Del(ctx, s.sessionKey(token))
	}

	pipe.Del(ctx, s.userSessionsKey(userID))

	_, err = pipe.Exec(ctx)
	if err != nil {
		log.Error("failed to delete user sessions from redis", err, logger.Field{Key: "user_id", Value: userID})
		return fmt.Errorf("failed to delete user sessions: %w", err)
	}

	log.Info("all sessions deleted successfully", logger.Field{Key: "user_id", Value: userID}, logger.Field{Key: "count", Value: len(refreshTokens)})
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
	var session domain.Session

	data, err := s.redis.Get(ctx, s.sessionKey(refreshToken)).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return session, domain.ErrSessionNotFound
		}
		return session, fmt.Errorf("failed to get session: %w", err)
	}

	if err := json.Unmarshal([]byte(data), &session); err != nil {
		return session, fmt.Errorf("failed to unmarshal session: %w", err)
	}

	return session, nil
}

func (s *Session) GetByUserId(ctx context.Context, userID string) ([]domain.Session, error) {
	refreshTokens, err := s.redis.SMembers(ctx, s.userSessionsKey(userID)).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get user sessions: %w", err)
	}

	sessions := make([]domain.Session, 0, len(refreshTokens))

	for _, token := range refreshTokens {
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
	log := s.logger.WithContext(ctx)
	log.Info("replacing session", logger.Field{Key: "user_id", Value: userID})

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
		if err.Error() == "session not found" {
			log.Warn("session not found for replacement", logger.Field{Key: "user_id", Value: userID})
			return domain.ErrSessionNotFound
		}
		log.Error("failed to replace session", err, logger.Field{Key: "user_id", Value: userID})
		return fmt.Errorf("failed to replace session: %w", err)
	}

	log.Info("session replaced successfully", logger.Field{Key: "user_id", Value: userID})
	return nil
}
