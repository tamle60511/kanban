package utils

import (
	"fmt"
	"strings"

	"github.com/go-playground/validator/v10"
)

var validate = validator.New()

// ValidateStruct validates a struct against its validation tags
func ValidateStruct(s interface{}) error {
	if err := validate.Struct(s); err != nil {
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			errorMessages := make([]string, 0, len(validationErrors))
			for _, e := range validationErrors {
				errorMessages = append(errorMessages, formatValidationError(e))
			}
			return fmt.Errorf("validation failed: %s", strings.Join(errorMessages, "; "))
		}
		return err
	}
	return nil
}

// formatValidationError formats a validation error into a human-readable message
func formatValidationError(e validator.FieldError) string {
	field := e.Field()
	tag := e.Tag()

	switch tag {
	case "required":
		return fmt.Sprintf("%s is required", field)
	case "email":
		return fmt.Sprintf("%s must be a valid email address", field)
	case "min":
		return fmt.Sprintf("%s must be at least %s", field, e.Param())
	case "max":
		return fmt.Sprintf("%s must be at most %s", field, e.Param())
	case "len":
		return fmt.Sprintf("%s must be exactly %s characters long", field, e.Param())
	case "eqfield":
		return fmt.Sprintf("%s must be equal to %s", field, e.Param())
	default:
		return fmt.Sprintf("%s failed on the '%s' validation", field, tag)
	}
}
