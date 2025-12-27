package validators

import (
	"errors"
	"fmt"
	"strings"

	"github.com/go-playground/validator/v10"
)

type ValidationMessage struct {
	Field   string `json:"field"`
	Message string `json:"message"`
	Tag     string `json:"tag,omitempty"`
	Value   string `json:"value,omitempty"`
}

func GetValidationMessage(fe validator.FieldError) ValidationMessage {
	fieldName := fe.Field()

	msg := ValidationMessage{
		Field: fieldName,
		Tag:   fe.Tag(),
	}

	switch fe.Tag() {
	case "required":
		msg.Message = fmt.Sprintf("%s is required", getFieldDisplayName(fieldName))

	case "email":
		msg.Message = "Invalid email address format"

	case "min":
		if fe.Type().String() == "string" {
			msg.Message = fmt.Sprintf("%s must be at least %s characters long", getFieldDisplayName(fieldName), fe.Param())
		} else {
			msg.Message = fmt.Sprintf("%s must be at least %s", getFieldDisplayName(fieldName), fe.Param())
		}

	case "max":
		if fe.Type().String() == "string" {
			msg.Message = fmt.Sprintf("%s must not exceed %s characters", getFieldDisplayName(fieldName), fe.Param())
		} else {
			msg.Message = fmt.Sprintf("%s must not exceed %s", getFieldDisplayName(fieldName), fe.Param())
		}

	case "phone":
		msg.Message = "Invalid phone number format. Use E.164 format (e.g., +79991234567)"

	case "strong_password":
		msg.Message = "Password must be at least 8 characters and contain letters, digits, and special characters"

	case "alphanum":
		msg.Message = fmt.Sprintf("%s must contain only alphanumeric characters", getFieldDisplayName(fieldName))

	case "alpha":
		msg.Message = fmt.Sprintf("%s must contain only alphabetic characters", getFieldDisplayName(fieldName))

	case "numeric":
		msg.Message = fmt.Sprintf("%s must be a valid number", getFieldDisplayName(fieldName))

	case "len":
		msg.Message = fmt.Sprintf("%s must be exactly %s characters long", getFieldDisplayName(fieldName), fe.Param())

	case "oneof":
		msg.Message = fmt.Sprintf("%s must be one of: %s", getFieldDisplayName(fieldName), fe.Param())

	case "url":
		msg.Message = "Invalid URL format"

	case "uuid":
		msg.Message = "Invalid UUID format"

	default:
		msg.Message = fmt.Sprintf("%s failed validation (%s)", getFieldDisplayName(fieldName), fe.Tag())
	}

	return msg
}

func getFieldDisplayName(fieldName string) string {
	var result strings.Builder
	for i, r := range fieldName {
		if i > 0 && r >= 'A' && r <= 'Z' {
			result.WriteRune(' ')
		}
		if i == 0 {
			result.WriteRune(r - ('a' - 'A'))
		} else {
			result.WriteRune(r)
		}
	}

	return result.String()
}

// FormatValidationErrors конвертирует ошибки валидации в понятные сообщения
// Больше не требует jsonTagMap - маппинг настроен в validator через RegisterTagNameFunc
func FormatValidationErrors(err error) []ValidationMessage {
	var validationErrors validator.ValidationErrors
	ok := errors.As(err, &validationErrors)
	if !ok {
		return []ValidationMessage{
			{
				Field:   "unknown",
				Message: "Validation failed",
				Tag:     "unknown",
			},
		}
	}

	messages := make([]ValidationMessage, 0, len(validationErrors))
	for _, fe := range validationErrors {
		msg := GetValidationMessage(fe)
		messages = append(messages, msg)
	}

	return messages
}
