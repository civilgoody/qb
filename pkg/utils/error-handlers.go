package utils

import (
	"errors"
	"fmt"
	"strconv"
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

// GetIDFromParam extracts an integer ID from a Gin context parameter.
// It returns the ID and a *BusinessError if parsing fails.
func GetIDFromParam(c *gin.Context, paramName string) (int, error) {
	idStr := c.Param(paramName)
	id, err := strconv.Atoi(idStr)
	if err != nil {
		return 0, NewValidationError(fmt.Sprintf("%s must be a valid integer", paramName))
	}
	return id, nil
}

// HandleDelete processes the result of a GORM delete operation.
func HandleDelete(c *gin.Context, result *gorm.DB, resourceType string, id int) {
	if result.Error != nil {
		HandleError(c, ErrDatabase)
		return
	}

	if result.RowsAffected == 0 {
		HandleError(c, ErrNotFound)
		return
	}

	SuccessResponse(c, gin.H{"message": fmt.Sprintf("%s with ID %d deleted successfully", resourceType, id)})
}
