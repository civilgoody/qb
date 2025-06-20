package utils

import (
	"errors"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"gorm.io/gorm"
)

// HandleValidationError processes validation errors in a standardized way
func HandleValidationError(c *gin.Context, err error) {
	if validationErrors, ok := err.(validator.ValidationErrors); ok {
		details := FormatValidationErrors(validationErrors)
		HandleError(c, NewValidationError(details))
		return
	}
	HandleError(c, NewValidationError(err.Error()))
}

// HandleDatabaseError processes database errors with common patterns
func HandleDatabaseError(c *gin.Context, err error) {
	if strings.Contains(err.Error(), "Duplicate entry") || strings.Contains(err.Error(), "unique constraint") {
		HandleError(c, ErrDuplicate)
		return
	}
	// Check if it's a GORM No Records Found error specifically
	if errors.Is(err, gorm.ErrRecordNotFound) {
		HandleError(c, ErrNotFound)
		return
	}
	HandleError(c, ErrDatabase)
}
