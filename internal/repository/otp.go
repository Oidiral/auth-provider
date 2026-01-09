package repository

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"time"

	"github.com/Oidiral/auth-provider/internal/domain"
	"github.com/Oidiral/auth-provider/pkg/logger"
	"github.com/redis/go-redis/v9"
)

type OtpRepository struct {
	redis      *redis.Client
	logger     logger.Logger
	ExpiresTTL time.Duration
}

func NewOtpRepository(redis *redis.Client, log logger.Logger, expiresTTL time.Duration) *OtpRepository {
	return &OtpRepository{
		redis:      redis,
		logger:     log,
		ExpiresTTL: expiresTTL,
	}
}

func generateNumericCode(length int) (string, error) {
	max := big.NewInt(1)
	for i := 0; i < length; i++ {
		max.Mul(max, big.NewInt(10))
	}
	n, err := rand.Int(rand.Reader, max)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%0*d", length, n), nil
}

func (o *OtpRepository) otpKey(userId string) string {
	return fmt.Sprintf("otp:%s", userId)
}

func (o *OtpRepository) Create(ctx context.Context, userId string) (string, error) {
	log := o.logger.WithContext(ctx)

	code, err := generateNumericCode(6)
	if err != nil {
		log.Error("failed to generate OTP code", err, logger.Field{Key: "user_id", Value: userId})
		return "", err
	}

	otp := domain.OTP{
		UserId:    userId,
		Code:      code,
		CreatedAt: time.Now(),
	}
	data, err := json.Marshal(otp)
	if err != nil {
		log.Error("failed to marshal OTP", err, logger.Field{Key: "user_id", Value: userId})
		return "", err
	}

	err = o.redis.Set(ctx, o.otpKey(userId), data, o.ExpiresTTL).Err()
	if err != nil {
		log.Error("failed to save OTP to redis", err, logger.Field{Key: "user_id", Value: userId})
		return "", err
	}

	return code, nil
}

func (o *OtpRepository) Delete(ctx context.Context, userId string) error {
	log := o.logger.WithContext(ctx)

	err := o.redis.Del(ctx, o.otpKey(userId)).Err()
	if err != nil {
		log.Error("failed to delete OTP from redis", err, logger.Field{Key: "user_id", Value: userId})
		return err
	}

	return nil
}

func (o *OtpRepository) Get(ctx context.Context, userId string) (domain.OTP, error) {
	log := o.logger.WithContext(ctx)

	var otp domain.OTP
	data, err := o.redis.Get(ctx, o.otpKey(userId)).Bytes()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return otp, domain.ErrOTPNotFound
		}
		log.Error("failed to get OTP from redis", err, logger.Field{Key: "user_id", Value: userId})
		return otp, domain.ErrOTPInternalServerError
	}
	err = json.Unmarshal(data, &otp)
	if err != nil {
		log.Error("failed to unmarshal OTP", err, logger.Field{Key: "user_id", Value: userId})
		return otp, err
	}

	return otp, nil
}

func (o *OtpRepository) Verify(ctx context.Context, userId, code string) error {
	log := o.logger.WithContext(ctx)

	script := redis.NewScript(`
		local otpData = redis.call('GET', KEYS[1])
		if not otpData then
			return redis.error_reply('otp not found')
		end

		local otp = cjson.decode(otpData)

		if otp.code ~= ARGV[1] then
			return redis.error_reply('otp invalid')
		end

		redis.call('DEL', KEYS[1])

		return 1
	`)

	err := script.Run(ctx, o.redis,
		[]string{o.otpKey(userId)},
		code,
	).Err()

	if err != nil {
		if err.Error() == "otp not found" {
			return domain.ErrOTPNotFound
		}
		if err.Error() == "otp invalid" {
			log.Warn("invalid OTP code provided", logger.Field{Key: "user_id", Value: userId})
			return domain.ErrOTPInvalid
		}
		log.Error("failed to verify OTP", err, logger.Field{Key: "user_id", Value: userId})
		return fmt.Errorf("failed to verify otp: %w", err)
	}

	return nil
}
