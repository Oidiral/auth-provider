package domain

import "errors"

var (
	// Session errors
	ErrSessionNotFound      = errors.New("session not found")
	ErrSessionExpired       = errors.New("session expired")
	ErrSessionInvalid       = errors.New("session invalid")
	ErrSessionAlreadyExists = errors.New("session already exists")

	// User errors
	ErrUserNotFound           = errors.New("user not found")
	ErrUserAlreadyExists      = errors.New("user already exists")
	ErrUserInvalidCredentials = errors.New("invalid credentials")
	ErrUserHasDependencies    = errors.New("user has dependencies and cannot be deleted")

	// Token errors
	ErrTokenInvalid          = errors.New("token invalid")
	ErrTokenExpired          = errors.New("token expired")
	ErrTokenMalformed        = errors.New("token malformed")
	ErrTokenSignatureInvalid = errors.New("token signature invalid")

	// OTP errors
	ErrOTPInvalid             = errors.New("otp invalid")
	ErrOTPExpired             = errors.New("otp expired")
	ErrOTPAlreadyUsed         = errors.New("otp already used")
	ErrOTPMaxAttemptsExceeded = errors.New("otp max attempts exceeded")

	// General errors
	ErrInternalServer = errors.New("internal server error")
	ErrBadRequest     = errors.New("bad request")
	ErrUnauthorized   = errors.New("unauthorized")
	ErrForbidden      = errors.New("forbidden")
	ErrNotFound       = errors.New("not found")
	ErrConflict       = errors.New("conflict")

	// Validation errors
	ErrInvalidInput       = errors.New("invalid input")
	ErrInvalidEmail       = errors.New("invalid email")
	ErrInvalidPassword    = errors.New("invalid password")
	ErrInvalidPhoneNumber = errors.New("invalid phone number")
	ErrInvalidUserID      = errors.New("invalid user id")
)
