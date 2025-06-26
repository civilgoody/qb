package models

import "fmt"

// BusinessError represents a structured business logic error
type BusinessError struct {
	Code    int
	Message string
	Details interface{}
}

func (e *BusinessError) Error() string {
	if e.Details != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Details)
	}
	return e.Message
}

// Common business errors - moved from utils
var (
	ErrValidation    = &BusinessError{Code: 400, Message: "Validation failed"}
	ErrUnauthorized  = &BusinessError{Code: 401, Message: "Unauthorized"}
	ErrForbidden     = &BusinessError{Code: 403, Message: "Forbidden"}
	ErrNotFound      = &BusinessError{Code: 404, Message: "Resource not found"}
	ErrDuplicate     = &BusinessError{Code: 409, Message: "Resource already exists"}
	ErrForeignKey    = &BusinessError{Code: 400, Message: "Referenced resource does not exist"}
	ErrDatabase      = &BusinessError{Code: 500, Message: "Database error"}
	ErrInternal      = &BusinessError{Code: 500, Message: "Internal server error"}
	ErrUploadFailed  = &BusinessError{Code: 500, Message: "File upload failed"}
	ErrNetworkIssue  = &BusinessError{Code: 503, Message: "Network connectivity issue"}
	ErrPartialUpload = &BusinessError{Code: 207, Message: "Some files uploaded successfully, others failed"}
)

// Error constructors - moved from utils
func NewValidationError(details interface{}) error {
	return &BusinessError{
		Code:    400,
		Message: "Validation failed",
		Details: details,
	}
}

func NewUploadError(details interface{}) error {
	return &BusinessError{
		Code:    500,
		Message: "File upload failed",
		Details: details,
	}
}

func NewNetworkError(details interface{}) error {
	return &BusinessError{
		Code:    503,
		Message: "Network connectivity issue",
		Details: details,
	}
}

func NewPartialUploadError(details interface{}) error {
	return &BusinessError{
		Code:    207,
		Message: "Some files uploaded successfully, others failed",
		Details: details,
	}
} 
