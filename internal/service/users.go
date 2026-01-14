package service

import (
	"context"
	"fmt"
	"time"

	"github.com/Oidiral/auth-provider/internal/domain"
	"github.com/Oidiral/auth-provider/internal/repository"
	"github.com/Oidiral/auth-provider/pkg/auth"
	"github.com/Oidiral/auth-provider/pkg/logger"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
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
	tracer          trace.Tracer
}

func NewUserService(repository repository.Users, uowFactory repository.UoWFactory, tokenManager auth.TokenManager, accessTokenTTL time.Duration, refreshTokenTTL time.Duration, sessionManager repository.Sessions, otpManager repository.OtpCodesRepository, log logger.Logger, tracer trace.Tracer) *UserService {
	return &UserService{
		repository:      repository,
		uowFactory:      uowFactory,
		tokenManager:    tokenManager,
		accessTokenTTL:  accessTokenTTL,
		refreshTokenTTL: refreshTokenTTL,
		sessionManager:  sessionManager,
		otpManager:      otpManager,
		logger:          log,
		tracer:          tracer,
	}
}

func (u *UserService) SignUp(ctx context.Context, input UserSignUpInput) error {
	ctx, span := u.tracer.Start(ctx, "UserService.SignUp")
	defer span.End()

	span.SetAttributes(attribute.String("user.email", input.Email))
	log := u.logger.WithContext(ctx)

	hashPassword, err := hash(input.Password)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to hash password")
		log.Error("failed to hash password", err)
		return err
	}

	uow, err := u.uowFactory.Begin(ctx)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to begin transaction")
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
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to create user")
		return err
	}
	span.SetAttributes(attribute.String("user.id", userId))

	role, err := uow.Roles().GetByName(ctx, domain.RoleUser)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to get default role")
		log.Error("failed to get default role", err)
		return err
	}

	err = uow.Roles().AssignToUser(ctx, userId, role.ID)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to assign role")
		log.Error("failed to assign role to user", err, logger.Field{Key: "user_id", Value: userId})
		return err
	}

	if err := uow.Commit(); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to commit transaction")
		log.Error("failed to commit transaction", err)
		return err
	}

	// TODO: В будущем реализовать отправку otp пользователю по sms или по email
	_, err = u.otpManager.Create(ctx, userId)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to create OTP")
		return err
	}

	return nil
}

func (u *UserService) SignIn(ctx context.Context, input UserSignInInput) (Tokens, error) {
	ctx, span := u.tracer.Start(ctx, "UserService.SignIn")
	defer span.End()

	span.SetAttributes(attribute.String("user.email", input.Email))
	log := u.logger.WithContext(ctx)

	uow, err := u.uowFactory.Begin(ctx)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to begin transaction")
		log.Error("failed to begin transaction", err)
		return Tokens{}, err
	}
	defer func() { _ = uow.Rollback() }()

	user, err := uow.Users().GetByEmail(ctx, input.Email)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to get user")
		return Tokens{}, err
	}
	span.SetAttributes(attribute.String("user.id", user.ID))

	if !user.IsVerified {
		span.SetStatus(codes.Error, "user not verified")
		log.Warn("unverified user signin attempt", logger.Field{Key: "user_id", Value: user.ID})
		return Tokens{}, domain.ErrUserNotActivated
	}

	roles, err := uow.Roles().GetUserRoles(ctx, user.ID)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to get user roles")
		log.Error("failed to get user roles", err, logger.Field{Key: "user_id", Value: user.ID})
		return Tokens{}, err
	}

	err = compare(input.Password, user.PasswordHash)
	if err != nil {
		span.SetStatus(codes.Error, "invalid credentials")
		log.Warn("invalid credentials", logger.Field{Key: "user_id", Value: user.ID})
		return Tokens{}, domain.ErrUserInvalidCredentials
	}

	access, err := u.tokenManager.NewJWT(user.ID, u.accessTokenTTL, roles)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to generate access token")
		log.Error("failed to generate access token", err, logger.Field{Key: "user_id", Value: user.ID})
		return Tokens{}, err
	}

	refresh, err := u.tokenManager.NewRefreshToken(u.refreshTokenTTL)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to generate refresh token")
		log.Error("failed to generate refresh token", err)
		return Tokens{}, err
	}

	err = u.sessionManager.Create(ctx, refresh, user.ID)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to create session")
		return Tokens{}, err
	}

	return Tokens{
		AccessToken:  access,
		RefreshToken: refresh,
	}, nil
}

func (u *UserService) Refresh(ctx context.Context, refreshToken string) (Tokens, error) {
	ctx, span := u.tracer.Start(ctx, "UserService.Refresh")
	defer span.End()

	log := u.logger.WithContext(ctx)

	if refreshToken == "" {
		span.SetStatus(codes.Error, "empty refresh token")
		return Tokens{}, domain.ErrTokenInvalid
	}

	session, err := u.sessionManager.Get(ctx, refreshToken)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to get session")
		return Tokens{}, err
	}
	span.SetAttributes(attribute.String("user.id", session.UserID))

	uow, err := u.uowFactory.Begin(ctx)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to begin transaction")
		log.Error("failed to begin transaction", err)
		return Tokens{}, err
	}
	defer func() { _ = uow.Rollback() }()

	roles, err := uow.Roles().GetUserRoles(ctx, session.UserID)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to get user roles")
		return Tokens{}, err
	}

	access, err := u.tokenManager.NewJWT(session.UserID, u.accessTokenTTL, roles)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to generate access token")
		log.Error("failed to generate new access token", err, logger.Field{Key: "user_id", Value: session.UserID})
		return Tokens{}, err
	}

	newRefreshToken, err := u.tokenManager.NewRefreshToken(u.refreshTokenTTL)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to generate refresh token")
		log.Error("failed to generate new refresh token", err, logger.Field{Key: "user_id", Value: session.UserID})
		return Tokens{}, err
	}

	err = u.sessionManager.Replace(ctx, refreshToken, newRefreshToken, session.UserID)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to replace session")
		return Tokens{}, err
	}

	return Tokens{
		AccessToken:  access,
		RefreshToken: newRefreshToken,
	}, nil
}

func (u *UserService) Verify(ctx context.Context, userId string, otpCode string) error {
	ctx, span := u.tracer.Start(ctx, "UserService.Verify")
	defer span.End()

	span.SetAttributes(attribute.String("user.id", userId))

	user, err := u.repository.Get(ctx, userId)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to get user")
		return err
	}
	if user.IsVerified {
		span.SetStatus(codes.Error, "user already verified")
		return domain.ErrOTPAlreadyActive
	}

	err = u.otpManager.Verify(ctx, userId, otpCode)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "OTP verification failed")
		return err
	}

	return nil
}

func (u *UserService) OtpRetrySend(ctx context.Context, userId string) error {
	ctx, span := u.tracer.Start(ctx, "UserService.OtpRetrySend")
	defer span.End()

	span.SetAttributes(attribute.String("user.id", userId))

	if userId == "" {
		span.SetStatus(codes.Error, "invalid user ID")
		return domain.ErrInvalidUserID
	}
	// TODO: В будущем реализовать отправку otp пользователю по sms или по email
	_, err := u.otpManager.Create(ctx, userId)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to create OTP")
	}
	return err
}

func (u *UserService) ValidateToken(ctx context.Context, token string) (userId string, role []string, err error) {
	ctx, span := u.tracer.Start(ctx, "UserService.ValidateToken")
	defer span.End()

	userId, role, err = u.tokenManager.Parse(token)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "invalid token")
		return "", nil, err
	}

	span.SetAttributes(attribute.String("user.id", userId))
	return userId, role, nil
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
