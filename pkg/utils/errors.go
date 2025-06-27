package utils

import (
	"fmt"
	"strings"

	"github.com/gin-gonic/gin"
)

// APIResponse defines the standard response structure
type APIResponse struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
	Error   interface{} `json:"error,omitempty"`
}

// BusinessError represents a structured business logic error
type BusinessError struct {
	Code    int
	Message string
	Details interface{} // Details holds structured error information, e.g., validation errors map
}

func (e *BusinessError) Error() string {
	// Return a general error message for logs, not for client display if Details exist
	if e.Details != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Details)
	}
	return e.Message
}

// Common business errors
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

// SuccessResponse sends a standardized success response
func SuccessResponse(c *gin.Context, data interface{}) {
	c.JSON(200, APIResponse{
		Code:    200,
		Message: "success",
		Data:    data,
	})
}

// ErrorResponse sends a standardized error response
func ErrorResponse(c *gin.Context, httpCode int, businessErr *BusinessError, details interface{}) {
	c.JSON(httpCode, APIResponse{
		Code:    businessErr.Code,
		Message: businessErr.Message,
		Error:   details,
	})
}

// HandleError processes different error types and sends appropriate responses
func HandleError(c *gin.Context, err error) {
	if be, ok := err.(*BusinessError); ok {
		// Business error with appropriate HTTP status
		httpCode := be.Code
		ErrorResponse(c, httpCode, be, be.Details)
		return
	}

	// Unknown error defaults to internal server error
	ErrorResponse(c, 500, ErrInternal, err.Error())
}

// NewValidationError creates a validation error with specific details
func NewValidationError(details interface{}) error {
	return &BusinessError{
		Code:    400,
		Message: "Validation failed", // User-friendly message for the client
		Details: details,             // The actual validation error map or string
	}
}

// NewUploadError creates an upload error with specific details
func NewUploadError(details interface{}) error {
	return &BusinessError{
		Code:    500,
		Message: "File upload failed",
		Details: details,
	}
}

// NewNetworkError creates a network error with specific details
func NewNetworkError(details interface{}) error {
	return &BusinessError{
		Code:    503,
		Message: "Network connectivity issue",
		Details: details,
	}
}

// NewPartialUploadError creates a partial upload error with specific details
func NewPartialUploadError(details interface{}) error {
	return &BusinessError{
		Code:    207,
		Message: "Some files uploaded successfully, others failed",
		Details: details,
	}
}

// isNetworkError checks if an error message indicates a network-related issue
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
