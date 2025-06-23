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

func BindAndValidate[T any](c *gin.Context, input *T) bool {
    if err := c.ShouldBindJSON(input); err != nil {
        HandleValidationError(c, err)
        return true // handled
    }
    
    if err := Validator.Struct(input); err != nil {
        HandleValidationError(c, err)
        return true // handled
    }
    
    return false // not handled
}

// MySQL error code mappings
var mysqlErrorMap = map[uint16]*BusinessError{
	1452: ErrForeignKey, // ER_NO_REFERENCED_ROW_2 - foreign key constraint fails
	1451: {Code: 400, Message: "Cannot delete resource: it is referenced by other records"}, // ER_ROW_IS_REFERENCED_2
	1062: ErrDuplicate, // ER_DUP_ENTRY - duplicate entry for unique key
	1406: {Code: 400, Message: "Data too long for one or more fields"}, // ER_DATA_TOO_LONG
}

// HandleDatabaseError processes database errors with proper error code detection
func HandleDatabaseError(c *gin.Context, err error) {
	// Handle MySQL-specific errors using map lookup
	if mysqlErr, ok := err.(*mysql.MySQLError); ok {
		if businessErr, exists := mysqlErrorMap[mysqlErr.Number]; exists {
			HandleError(c, businessErr)
			return
		}
	}
	
	// Handle GORM-specific errors
	if errors.Is(err, gorm.ErrRecordNotFound) {
		HandleError(c, ErrNotFound)
		return
	}
	
	// Log unhandled database error for debugging
	fmt.Printf("Unhandled database error: %v\n", err)
	HandleError(c, ErrDatabase)
}

// HandleDatabaseErrorWithContext processes database errors with specific resource context
func HandleDatabaseErrorWithContext(c *gin.Context, err error, resourceContext string) {
	// Handle MySQL-specific errors using map lookup
	if mysqlErr, ok := err.(*mysql.MySQLError); ok {
		if businessErr, exists := mysqlErrorMap[mysqlErr.Number]; exists {
			HandleError(c, businessErr)
			return
		}
	}
	
	// Handle GORM-specific errors with context
	if errors.Is(err, gorm.ErrRecordNotFound) {
		HandleError(c, &BusinessError{
			Code:    404,
			Message: fmt.Sprintf("%s not found", resourceContext),
		})
		return
	}
	
	// Log unhandled database error for debugging
	fmt.Printf("Unhandled database error for %s: %v\n", resourceContext, err)
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

	result := db.Delete(model, "id = ?", id)
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

// GetCourseFilterParams extracts and validates dept, level, and semester parameters
// Returns the validated parameters and a boolean indicating if an error was handled
func GetCourseFilterParams(c *gin.Context) (dept string, level int, semester int, handled bool) {
    dept = c.Param("dept")
    
    // Validate and get level
    level, handled = GetAndHandleID(c, "level")
    if handled {
        return "", 0, 0, true
    }
    
    // Validate and get semester
    semester, handled = GetAndHandleID(c, "semester")
    if handled {
        return "", 0, 0, true
    }
    
    // Additional validation for level (100, 200, 300, 400, 500)
    if level < 100 || level > 500 || level%100 != 0 {
        HandleError(c, NewValidationError("Level must be one of: 100, 200, 300, 400, 500"))
        return "", 0, 0, true
    }
    
    // Additional validation for semester (1 or 2)
    if semester < 1 || semester > 2 {
        HandleError(c, NewValidationError("Semester must be 1 or 2"))
        return "", 0, 0, true
    }
    
    return dept, level, semester, false
}
