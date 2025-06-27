package services

import (
	"errors"
	"fmt"
	"qb/pkg/models"
	"strconv"
	"github.com/go-playground/validator/v10"
	"github.com/go-sql-driver/mysql"
	"gorm.io/gorm"
)

type ErrorService struct{}

func NewErrorService() *ErrorService {
	return &ErrorService{}
}

// MySQL error code mappings - moved from utils
var mysqlErrorMap = map[uint16]*models.BusinessError{
	1452: models.ErrForeignKey,
	1451: {Code: 400, Message: "Cannot delete resource: it is referenced by other records"},
	1062: models.ErrDuplicate,
	1406: {Code: 400, Message: "Data too long for one or more fields"},
}

// ProcessDatabaseError handles database errors
func (s *ErrorService) Db(err error, resourceContext ...string) error {
	// Handle MySQL-specific errors
	if mysqlErr, ok := err.(*mysql.MySQLError); ok {
		if businessErr, exists := mysqlErrorMap[mysqlErr.Number]; exists {
			return businessErr
		}
	}
	
	// Handle GORM-specific errors
	if errors.Is(err, gorm.ErrRecordNotFound) {
		if len(resourceContext) > 0 {
			return &models.BusinessError{
				Code:    404,
				Message: fmt.Sprintf("%s not found", resourceContext[0]),
			}
		}
		return models.ErrNotFound
	}
	
	// Log unhandled database error
	fmt.Printf("Unhandled database error: %v\n", err)
	return models.ErrDatabase
}

// ProcessValidationError handles validation errors
func (s *ErrorService) Invalid(err interface{}) error {
	if validationErrors, ok := err.(validator.ValidationErrors); ok {
		details := s.formatValidationErrors(validationErrors)
		return models.NewValidationError(details)
	} else if err, ok := err.(error); ok {
		return models.NewValidationError(err.Error())
	}
	return models.NewValidationError(err)
}

// GetIDFromParam extracts integer ID from URL param
func (s *ErrorService) GetIntId(idStr string, paramName string) (int, error) {
	id, err := strconv.Atoi(idStr)
	if err != nil {
		return 0, models.NewValidationError(paramName + " must be a valid integer")
	}
	return id, nil
}

// formatValidationErrors formats validator errors - ultra simple
func (s *ErrorService) formatValidationErrors(errors validator.ValidationErrors) map[string]string {
	validationMessages := map[string]string{
		"required": "is required",
		"email":    "must be a valid email",
		"min":      "is too short",
		"max":      "is too long", 
		"len":      "has invalid length",
		"oneof":    "has invalid value",
		"url":      "must be a valid URL",
	}

	errorMap := make(map[string]string)
	for _, err := range errors {
		field := err.Field()
		tag := err.Tag()
		
		if message, exists := validationMessages[tag]; exists {
			errorMap[field] = fmt.Sprintf("%s %s", field, message)
		} else {
			errorMap[field] = fmt.Sprintf("%s is invalid", field)
		}
	}
	
	return errorMap
} 
