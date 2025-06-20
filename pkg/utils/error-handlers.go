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

// GetAndHandleID extracts an integer ID from a Gin context parameter and handles errors.
// It returns the ID and a boolean indicating if an error was handled.
func GetAndHandleID(c *gin.Context, paramName string) (id int, handled bool) {
	id, err := GetIDFromParam(c, paramName)
	if err != nil {
		HandleError(c, err)
		return 0, true // Error handled, return 0 and true
	}
	return id, false // No error, return ID and false
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

// DeleteResourceByID handles the common logic for deleting a resource by ID.
// It returns true if an error occurred and was handled, false otherwise.
func DeleteResourceByID(c *gin.Context, db *gorm.DB, paramName, resourceType string, model interface{}) bool {
	id, err := GetIDFromParam(c, paramName)
	if err != nil {
		HandleError(c, err)
		return true
	}

	result := db.Delete(model, id)
	HandleDelete(c, result, resourceType, id)

	return false // No error occurred
}

// HandleCreateResource handles the common logic for creating a resource:
// binding JSON, validating, and creating in the database.
// It returns true if an error occurred and was handled, false otherwise.
func HandleCreateResource(c *gin.Context, db *gorm.DB, validate *validator.Validate, resource interface{}) bool {
	// Bind JSON payload to the resource struct
	if err := c.ShouldBindJSON(resource); err != nil {
		HandleValidationError(c, err)
		return true
	}

	// Validate the struct based on 'validate' tags
	if err := validate.Struct(resource); err != nil {
		HandleValidationError(c, err)
		return true
	}

	// Attempt to create the resource in the database
	if err := db.Create(resource).Error; err != nil {
		HandleDatabaseError(c, err)
		return true
	}

	SuccessResponse(c, resource) // Success response moved here
	return false // No error occurred
}

// HandleGetResources handles the common logic for retrieving a collection of resources.
// It returns true if an error occurred and was handled, false otherwise.
func HandleGetResources(c *gin.Context, db *gorm.DB, resources interface{}) bool {
	if err := db.Find(resources).Error; err != nil {
		HandleError(c, ErrDatabase)
		return true
	}
	SuccessResponse(c, resources)
	return false // No error occurred
}
