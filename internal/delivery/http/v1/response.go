package v1

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/Oidiral/auth-provider/internal/domain"
	apiv1 "github.com/Oidiral/auth-provider/internal/generated/api/v1"
	validators "github.com/Oidiral/auth-provider/pkg/validator"
	"github.com/go-playground/validator/v10"
)

func respondJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func respondError(w http.ResponseWriter, status int, message, details string) {
	var detailsPtr *string
	if details != "" {
		detailsPtr = &details
	}
	respondJSON(w, status, apiv1.ErrorResponse{
		Status:  status,
		Message: message,
		Details: detailsPtr,
	})
}

func respondValidationError(w http.ResponseWriter, err error) {
	var validationErrors validator.ValidationErrors
	if !errors.As(err, &validationErrors) {
		respondError(w, http.StatusBadRequest, "validation failed", "invalid input format")
		return
	}

	messages := validators.FormatValidationErrors(validationErrors)

	// Конвертируем validators.ValidationMessage в apiv1.ValidationMessage
	apiMessages := make([]apiv1.ValidationMessage, len(messages))
	for i, msg := range messages {
		var tagPtr, valuePtr *string
		if msg.Tag != "" {
			tagPtr = &msg.Tag
		}
		if msg.Value != "" {
			valuePtr = &msg.Value
		}

		apiMessages[i] = apiv1.ValidationMessage{
			Field:   msg.Field,
			Message: msg.Message,
			Tag:     tagPtr,
			Value:   valuePtr,
		}
	}

	response := apiv1.ValidationErrorResponse{
		Status:  http.StatusBadRequest,
		Message: "validation failed",
		Errors:  apiMessages,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusBadRequest)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "internal server error", http.StatusInternalServerError)
	}
}

func respondServiceError(w http.ResponseWriter, err error) {
	code := http.StatusInternalServerError
	message := "internal server error"
	details := ""

	switch {
	case errors.Is(err, domain.ErrUserAlreadyExists):
		code = http.StatusConflict
		message = "user already exists"
		details = "a user with this email or username already exists"

	case errors.Is(err, domain.ErrUserNotFound):
		code = http.StatusNotFound
		message = "user not found"
		details = "no user found with the provided credentials"

	case errors.Is(err, domain.ErrUserInvalidCredentials):
		code = http.StatusUnauthorized
		message = "invalid credentials"
		details = "email or password is incorrect"

	case errors.Is(err, domain.ErrUserNotActivated):
		code = http.StatusForbidden
		message = "user not activated, please check your email"

	case errors.Is(err, domain.ErrSessionExpired):
		code = http.StatusUnauthorized
		message = "session expired"
		details = "your session has expired, please log in again"

	case errors.Is(err, domain.ErrSessionNotFound):
		code = http.StatusNotFound
		message = "session not found"
		details = "no active session found"

	case errors.Is(err, domain.ErrOTPNotFound):
		code = http.StatusNotFound
		message = "OTP not found"
		details = "no OTP code found for this user"

	case errors.Is(err, domain.ErrOTPInvalid):
		code = http.StatusBadRequest
		message = "invalid OTP"
		details = "the OTP code is incorrect"

	case errors.Is(err, domain.ErrOTPAlreadyUsed):
		code = http.StatusBadRequest
		message = "OTP already used"
		details = "this OTP code has already been used"

	case errors.Is(err, domain.ErrOTPAlreadyActive):
		code = http.StatusBadRequest
		message = "active OTP exists"
		details = "an active OTP code already exists, please wait or use the existing code"

	default:
		details = ""
	}

	respondError(w, code, message, details)
}
