package models

// ErrorResponse represents a standardized error response structure for API failures.
type ErrorResponse struct {
	Error string `json:"error"` // A descriptive error message
}

// ValidationErrorResponse represents a standardized validation error response structure.
type ValidationErrorResponse struct {
	ValidationErrors map[string]string `json:"validation_errors"` // Map of field names to error messages
} 
