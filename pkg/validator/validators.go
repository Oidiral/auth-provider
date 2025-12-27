package validators

import (
	"regexp"
	"unicode"

	"github.com/go-playground/validator/v10"
)

// Format: +[country code][subscriber number]
var phoneRegex = regexp.MustCompile(`^\+?[1-9]\d{9,14}$`)

func RegisterCustomValidators(v *validator.Validate) error {
	validators := map[string]validator.Func{
		"phone":           validatePhone,
		"strong_password": validateStrongPassword,
	}

	for name, fn := range validators {
		if err := v.RegisterValidation(name, fn); err != nil {
			return err
		}
	}

	return nil
}

func validatePhone(fl validator.FieldLevel) bool {
	phone := fl.Field().String()
	if phone == "" {
		return false
	}
	return phoneRegex.MatchString(phone)
}

// validateStrongPassword validates password strength.
// Requirements:
//   - Minimum 8 characters
//   - Contains at least one letter (any case)
//   - Contains at least one digit
//   - Contains at least one special character (punctuation or symbol)
func validateStrongPassword(fl validator.FieldLevel) bool {
	password := fl.Field().String()

	if len(password) < 8 {
		return false
	}

	var hasLetter, hasDigit, hasSpecial bool
	for _, r := range password {
		switch {
		case unicode.IsLetter(r):
			hasLetter = true
		case unicode.IsDigit(r):
			hasDigit = true
		case unicode.IsPunct(r) || unicode.IsSymbol(r):
			hasSpecial = true
		}
	}

	return hasLetter && hasDigit && hasSpecial
}
