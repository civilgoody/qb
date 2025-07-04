package utils

import (
	"fmt"
	"strings"

	"github.com/go-playground/validator/v10"
)

// Validator is the global instance of the validator.
var Validator = validator.New()

// FormatValidationErrors converts validator errors to user-friendly messages
func FormatValidationErrors(err error) map[string]string {
	errors := make(map[string]string)
	
	if validationErrors, ok := err.(validator.ValidationErrors); ok {
		for _, fieldErr := range validationErrors {
			fieldName := strings.ToLower(fieldErr.Field())
			
			switch fieldErr.Tag() {
			case "required":
				errors[fieldName] = fmt.Sprintf("%s is required", fieldName)
			case "oneof":
				errors[fieldName] = fmt.Sprintf("%s must be one of: %s", fieldName, fieldErr.Param())
			case "min":
				errors[fieldName] = fmt.Sprintf("%s must be at least %s characters", fieldName, fieldErr.Param())
			case "max":
				errors[fieldName] = fmt.Sprintf("%s must be at most %s characters", fieldName, fieldErr.Param())
			case "email":
				errors[fieldName] = fmt.Sprintf("%s must be a valid email address", fieldName)
			default:
				errors[fieldName] = fmt.Sprintf("%s is invalid", fieldName)
			}
		}
	}
	
	return errors
} 

func IsNetworkError(errorMsg string) bool {
	networkKeywords := []string{
		"TLS handshake timeout",
		"timeout",
		"network",
		"connection",
		"dial tcp",
		"no such host",
		"connection refused",
		"connection reset",
	}
	
	for _, keyword := range networkKeywords {
		if strings.Contains(strings.ToLower(errorMsg), strings.ToLower(keyword)) {
			return true
		}
	}
	return false
} 
