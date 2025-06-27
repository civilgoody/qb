package services

import (
	"qb/pkg/models"

	"github.com/go-playground/validator/v10"
	"gorm.io/gorm"
)

// GenericService handles basic CRUD operations for simple resources
type GenericService struct {
	db       *gorm.DB
	validate *validator.Validate
	err      *ErrorService
}

// NewGenericService creates a new Generic service instance
func NewGenericService(db *gorm.DB, validate *validator.Validate) *GenericService {
	return &GenericService{
		db:       db,
		validate: validate,
		err:      NewErrorService(),
	}
}

// CreateResource creates a new resource in the database
func (s *GenericService) CreateResource(resource interface{}) error {
	if err := s.validate.Struct(resource); err != nil {
		return s.err.Invalid(err)
	}

	if err := s.db.Create(resource).Error; err != nil {
		return s.err.Db(err)
	}

	return nil
}

// GetAllResources retrieves all resources of a given type
func (s *GenericService) GetAllResources(resources interface{}) error {
	if err := s.db.Find(resources).Error; err != nil {
		return s.err.Db(err)
	}
	return nil
}

// GetResourceByID gets a resource by integer ID
func (s *GenericService) GetResourceByID(id int, resource interface{}) error {
	if err := s.db.First(resource, id).Error; err != nil {
		return s.err.Db(err)
	}
	return nil
}

// GetResourceByStringID gets a resource by string ID
func (s *GenericService) GetResourceByStringID(id string, resource interface{}) error {
	if err := s.db.Where("id = ?", id).First(resource).Error; err != nil {
		return s.err.Db(err)
	}
	return nil
}

// DeleteResource deletes a resource by ID (handles both int and string IDs)
func (s *GenericService) DeleteResource(id interface{}, resource interface{}) error {
	var result *gorm.DB
		
	switch v := id.(type) {
	case int:
		result = s.db.Delete(resource, v)
	case string:
		result = s.db.Delete(resource, "id = ?", v)
	default:
		return s.err.Invalid("ID must be either int or string")
	}
	
	if result.Error != nil {
		return s.err.Db(result.Error)
	}
	
	if result.RowsAffected == 0 {
		return models.ErrNotFound
	}
	
	return nil
}

// GetDepartmentsWithFaculty retrieves all departments with their faculty preloaded
func (s *GenericService) GetDepartmentsWithFaculty() ([]models.Department, error) {
	var departments []models.Department
	if err := s.db.Preload("Faculty").Find(&departments).Error; err != nil {
		return nil, s.err.Db(err)
	}
	return departments, nil
} 
