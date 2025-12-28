package domain

import "errors"

var (
	// Session errors
	ErrSessionNotFound = errors.New("session not found")
	ErrSessionExpired  = errors.New("session expired")

	// User errors
	ErrUserNotFound           = errors.New("user not found")
	ErrUserAlreadyExists      = errors.New("user already exists")
	ErrUserInvalidCredentials = errors.New("invalid credentials")
	ErrUserHasDependencies    = errors.New("user has dependencies and cannot be deleted")

	// Token errors
	ErrTokenInvalid = errors.New("token invalid")

	// OTP errors
	ErrOTPInvalid            = errors.New("otp invalid")
	ErrOTPAlreadyUsed        = errors.New("otp already used")
	ErrOTPAlreadyActive      = errors.New("user already has an active otp")
	ErrOTPNotFound           = errors.New("otp not found")
	ErrUserNotActivated      = errors.New("user not activated")
	ErrOTPInteralServerError = errors.New("otp internal server error")

	// General errors
	ErrInternalServer = errors.New("internal server error")

	// Validation errors
	ErrInvalidInput  = errors.New("invalid input")
	ErrInvalidUserID = errors.New("invalid user id")

	// Roles errors
	ErrRoleNotFound        = errors.New("role not found")
	ErrRoleAlreadyAssigned = errors.New("role already assigned to user")
)
