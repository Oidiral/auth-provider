package auth

import (
	"errors"
	"fmt"
	"time"

	"github.com/Oidiral/auth-provider/internal/domain"
	"github.com/Oidiral/auth-provider/pkg/logger"
	"github.com/golang-jwt/jwt/v5"
)

type TokenManager interface {
	NewJWT(userId string, ttl time.Duration, roles []domain.Role) (string, error)
	Parse(accessToken string) (userId string, roles []string, err error)
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

func (m *Manager) NewJWT(userId string, ttl time.Duration, roles []domain.Role) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": userId,
		"roles":   roles,
		"exp":     time.Now().Add(ttl).Unix(),
		"type":    "access",
	})

	return token.SignedString([]byte(m.signingKey))
}

func (m *Manager) Parse(accessToken string) (userId string, roles []string, err error) {
	token, err := jwt.Parse(accessToken, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(m.signingKey), nil
	})
	if err != nil {
		return
	}

	if !token.Valid {
		err = errors.New("token is invalid")
		return
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		err = errors.New("invalid claims")
		return
	}

	userId, ok = claims["user_id"].(string)
	if !ok {
		err = errors.New("invalid user_id in token")
		return
	}
	rolesInterface, ok := claims["roles"].([]interface{})
	if !ok {
		err = errors.New("invalid roles in token")
		return
	}

	roles = make([]string, len(rolesInterface))
	for i, role := range rolesInterface {
		roleStr, ok := role.(string)
		if !ok {
			err = errors.New("invalid role type in token")
			return
		}
		roles[i] = roleStr
	}

	return
}

func (m *Manager) NewRefreshToken(ttl time.Duration) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"exp":  time.Now().Add(ttl).Unix(),
		"type": "refresh",
	})

	return token.SignedString([]byte(m.signingKey))
}
