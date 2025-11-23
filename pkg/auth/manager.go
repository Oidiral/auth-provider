package auth

import (
	"errors"
	"fmt"
	"time"

	"github.com/Oidiral/auth-provider/pkg/logger"
	"github.com/golang-jwt/jwt/v5"
)

type TokenManager interface {
	NewJWT(userId string, ttl time.Duration) (string, error)
	Parse(accessToken string) (string, error)
	NewRefreshToken(ttl time.Duration) (string, error)
}

type Manager struct {
	signingKey string
}

func NewManager(signingKey string, log logger.Logger) (*Manager, error) {
	if signingKey == "" {
		err := errors.New("empty signing key")
		log.Error("failed to initialize token manager", err)
		return nil, err
	}
	log.Info("token manager initialized successfully")
	return &Manager{
		signingKey: signingKey,
	}, nil
}

func (m *Manager) NewJWT(userId string, ttl time.Duration) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": userId,
		"exp":     time.Now().Add(ttl).Unix(),
		"type":    "access",
	})

	return token.SignedString([]byte(m.signingKey))
}

func (m *Manager) Parse(accessToken string) (string, error) {
	token, err := jwt.Parse(accessToken, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(m.signingKey), nil
	})
	if err != nil {
		return "", err
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return "", errors.New("invalid claims")
	}

	userId, ok := claims["user_id"].(string)
	if !ok || userId == "" {
		return "", errors.New("invalid or missing user_id in token")
	}
	return userId, nil
}

func (m *Manager) NewRefreshToken(ttl time.Duration) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"exp":  time.Now().Add(ttl).Unix(),
		"type": "refresh",
	})

	return token.SignedString([]byte(m.signingKey))
}
