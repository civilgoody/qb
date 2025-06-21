package utils

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/go-sql-driver/mysql"
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

// HandleDatabaseError processes database errors with proper error code detection
func HandleDatabaseError(c *gin.Context, err error) {
	// First, try to handle MySQL-specific errors by error code (more reliable)
	if mysqlErr, ok := err.(*mysql.MySQLError); ok {
		switch mysqlErr.Number {
		case 1452: // ER_NO_REFERENCED_ROW_2 - foreign key constraint fails (INSERT/UPDATE)
			HandleError(c, ErrForeignKey)
			return
		case 1451: // ER_ROW_IS_REFERENCED_2 - cannot delete parent row (foreign key constraint)
			HandleError(c, &BusinessError{
				Code: 400, 
				Message: "Cannot delete resource: it is referenced by other records",
			})
			return
		case 1062: // ER_DUP_ENTRY - duplicate entry for unique key
			HandleError(c, ErrDuplicate)
			return
		case 1406: // ER_DATA_TOO_LONG - data too long for column
			HandleError(c, &BusinessError{
				Code: 400,
				Message: "Data too long for one or more fields",
			})
			return
		}
	}
	
	// If it's not a MySQL error, or an unhandled MySQL error, fallback to generic checks.
	// Check if it's a GORM No Records Found error specifically
	if errors.Is(err, gorm.ErrRecordNotFound) {
		HandleError(c, ErrNotFound)
		return
	}
	
	// Log the unhandled database error for debugging
	fmt.Printf("Unhandled database error: %v\n", err)
	HandleError(c, ErrDatabase)
}

// GetIDFromParam extracts an integer ID from a Gin context parameter.
// It returns the ID and a *BusinessError if parsing fails.
func GetIDFromParam(c *gin.Context, paramName string                      ) (int, error) {
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
func HandleDelete(c *gin.Context, result *gorm.DB, resourceType string, id interface{}) {
	if result.Error != nil {
		HandleError(c, ErrDatabase)
		return
	}

	if result.RowsAffected == 0 {
		HandleError(c, ErrNotFound)
		return
	}

	SuccessResponse(c, gin.H{"message": fmt.Sprintf("%s with ID %v deleted successfully", resourceType, id)})
}

// DeleteResourceByID handles the common logic for deleting a resource by ID.
// It returns true if an error occurred and was handled, false otherwise.
func DeleteResourceByID(c *gin.Context, db *gorm.DB, paramName, resourceType string, model interface{}, isStrId ...bool) bool {
	id, err := GetIDFromParam(c, paramName)
	if err != nil {
		HandleError(c, err)
		return true
	}

	result := db.Delete(model, id)
	HandleDelete(c, result, resourceType, id)

	return false // No error occurred
}

func DeleteResourceByStringID(c *gin.Context, db *gorm.DB, paramName, resourceType string, model interface{}) bool {
	id := c.Param(paramName)

	result := db.Delete(model, "_id = ?", id)
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
