package service

import (
	"context"
	"fmt"
	"time"

	"github.com/Oidiral/auth-provider/internal/domain"
	"github.com/Oidiral/auth-provider/internal/repository"
	"github.com/Oidiral/auth-provider/pkg/auth"
	"github.com/Oidiral/auth-provider/pkg/logger"
	"golang.org/x/crypto/bcrypt"
)

type UserService struct {
	repository      repository.Users
	uowFactory      repository.UoWFactory
	tokenManager    auth.TokenManager
	sessionManager  repository.Sessions
	otpManager      repository.OtpCodesRepository
	accessTokenTTL  time.Duration
	refreshTokenTTL time.Duration
	logger          logger.Logger
}

func NewUserService(repository repository.Users, uowFactory repository.UoWFactory, tokenManager auth.TokenManager, accessTokenTTL time.Duration, refreshTokenTTL time.Duration, sessionManager repository.Sessions, otpManager repository.OtpCodesRepository, log logger.Logger) *UserService {
	return &UserService{
		repository:      repository,
		uowFactory:      uowFactory,
		tokenManager:    tokenManager,
		accessTokenTTL:  accessTokenTTL,
		refreshTokenTTL: refreshTokenTTL,
		sessionManager:  sessionManager,
		otpManager:      otpManager,
		logger:          log,
	}
}

func (u *UserService) SignUp(ctx context.Context, input UserSignUpInput) error {
	log := u.logger.WithContext(ctx)
	log.Info("user signup attempt", logger.Field{Key: "email", Value: input.Email})

	hashPassword, err := hash(input.Password)
	if err != nil {
		log.Error("failed to hash password", err)
		return err
	}

	uow, err := u.uowFactory.Begin(ctx)
	if err != nil {
		log.Error("failed to begin transaction", err)
		return err
	}
	defer func() { _ = uow.Rollback() }()

	var phone *string
	if input.Phone != "" {
		phone = &input.Phone
	}
	user := domain.User{
		Username:     input.Username,
		FirstName:    input.FirstName,
		LastName:     input.LastName,
		Phone:        phone,
		Email:        input.Email,
		PasswordHash: hashPassword,
	}

	userId, err := uow.Users().Create(ctx, user)
	if err != nil {
		log.Error("failed to create user", err, logger.Field{Key: "email", Value: input.Email})
		return err
	}

	role, err := uow.Roles().GetByName(ctx, domain.RoleUser)
	if err != nil {
		log.Error("failed to get default role", err)
		return err
	}

	err = uow.Roles().AssignToUser(ctx, userId, role.ID)
	if err != nil {
		log.Error("failed to assign role to user", err, logger.Field{Key: "user_id", Value: userId})
		return err
	}

	if err := uow.Commit(); err != nil {
		log.Error("failed to commit transaction", err)
		return err
	}

	err = u.otpManager.Create(ctx, userId)
	if err != nil {
		log.Error("failed to create OTP", err, logger.Field{Key: "user_id", Value: userId})
		return err
	}

	log.Info("user signed up successfully", logger.Field{Key: "email", Value: input.Email}, logger.Field{Key: "user_id", Value: userId})
	return nil
}

func (u *UserService) SignIn(ctx context.Context, input UserSignInInput) (Tokens, error) {
	log := u.logger.WithContext(ctx)
	log.Info("user signin attempt", logger.Field{Key: "email", Value: input.Email})

	user, err := u.repository.GetByEmail(ctx, input.Email)
	if err != nil {
		log.Warn("user not found", logger.Field{Key: "email", Value: input.Email})
		return Tokens{}, err
	}

	if !user.IsVerified {
		log.Warn("unverified user signin attempt", logger.Field{Key: "email", Value: input.Email}, logger.Field{Key: "user_id", Value: user.ID})
		return Tokens{}, domain.ErrUserNotActivated
	}

	err = compare(input.Password, user.PasswordHash)
	if err != nil {
		log.Warn("invalid credentials for user", logger.Field{Key: "email", Value: input.Email})
		return Tokens{}, domain.ErrUserInvalidCredentials
	}

	access, err := u.tokenManager.NewJWT(user.ID, u.accessTokenTTL)
	if err != nil {
		log.Error("failed to generate access token", err, logger.Field{Key: "user_id", Value: user.ID})
		return Tokens{}, err
	}

	refresh, err := u.tokenManager.NewRefreshToken(u.refreshTokenTTL)
	if err != nil {
		log.Error("failed to generate refresh token", err)
		return Tokens{}, err
	}

	err = u.sessionManager.Create(ctx, refresh, user.ID)
	if err != nil {
		log.Error("failed to create session", err, logger.Field{Key: "user_id", Value: user.ID})
		return Tokens{}, err
	}

	log.Info("user signed in successfully", logger.Field{Key: "email", Value: input.Email}, logger.Field{Key: "user_id", Value: user.ID})
	return Tokens{
		AccessToken:  access,
		RefreshToken: refresh,
	}, nil
}

func (u *UserService) Refresh(ctx context.Context, refreshToken string) (Tokens, error) {
	log := u.logger.WithContext(ctx)
	log.Info("token refresh attempt")

	if refreshToken == "" {
		log.Warn("empty refresh token provided")
		return Tokens{}, domain.ErrTokenInvalid
	}

	session, err := u.sessionManager.Get(ctx, refreshToken)
	if err != nil {
		log.Warn("invalid or expired refresh token")

		return Tokens{}, err
	}

	access, err := u.tokenManager.NewJWT(session.UserID, u.accessTokenTTL)
	if err != nil {
		log.Error("failed to generate new access token", err, logger.Field{Key: "user_id", Value: session.UserID})
		return Tokens{}, err
	}

	newRefreshToken, err := u.tokenManager.NewRefreshToken(u.refreshTokenTTL)
	if err != nil {
		log.Error("failed to generate new refresh token", err, logger.Field{Key: "user_id", Value: session.UserID})
		return Tokens{}, err
	}

	err = u.sessionManager.Replace(ctx, refreshToken, newRefreshToken, session.UserID)
	if err != nil {
		log.Error("failed to replace session", err, logger.Field{Key: "user_id", Value: session.UserID})
		return Tokens{}, err
	}

	log.Info("token refreshed successfully", logger.Field{Key: "user_id", Value: session.UserID})
	return Tokens{
		AccessToken:  access,
		RefreshToken: newRefreshToken,
	}, nil
}

func (u *UserService) Verify(ctx context.Context, userId string, otpCode string) error {
	log := u.logger.WithContext(ctx)
	log.Info("user verification attempt", logger.Field{Key: "user_id", Value: userId})

	user, err := u.repository.Get(ctx, userId)
	if err != nil {
		log.Error("failed to get user for verification", err, logger.Field{Key: "user_id", Value: userId})
		return err
	}
	if user.IsVerified {
		log.Warn("user already verified", logger.Field{Key: "user_id", Value: userId})
		return domain.ErrOTPAlreadyActive
	}

	err = u.otpManager.Verify(ctx, userId, otpCode)
	if err != nil {
		log.Warn("OTP verification failed", logger.Field{Key: "user_id", Value: userId})
		return err
	}

	log.Info("user verified successfully", logger.Field{Key: "user_id", Value: userId})
	return nil
}

func (u *UserService) OtpRetrySend(ctx context.Context, userId string) error {
	log := u.logger.WithContext(ctx)
	log.Info("OTP retry send attempt", logger.Field{Key: "user_id", Value: userId})

	if userId == "" {
		log.Warn("empty user id provided for OTP retry send")
		return domain.ErrInvalidUserID
	}
	return u.otpManager.Create(ctx, userId)
}

func hash(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("failed to hash password: %w", err)
	}
	return string(hash), nil
}

func compare(password, hash string) error {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	if err != nil {
		return fmt.Errorf("failed to compare password: %w", err)
	}
	return nil
}
