package services

import (
	"qb/pkg/models"

	"github.com/go-playground/validator/v10"
	"gorm.io/gorm"
)

// GenericService handles basic CRUD operations for simple resources
type GenericService struct {
	DB        *gorm.DB
	Validator *validator.Validate
}

// NewGenericService creates a new Generic service instance
func NewGenericService(db *gorm.DB, validator *validator.Validate) *GenericService {
	return &GenericService{
		DB:        db,
		Validator: validator,
	}
}

// Pure business logic methods

// CreateResource creates a new resource in the database
func (s *GenericService) CreateResource(resource interface{}) error {
	if err := s.Validator.Struct(resource); err != nil {
		return err
	}

	return s.DB.Create(resource).Error
}

// GetAllResources retrieves all resources of a given type
func (s *GenericService) GetAllResources(resources interface{}) error {
	return s.DB.Find(resources).Error
}

func (s *GenericService) GetResourceByID(id int, resource interface{}) error {
	return s.DB.First(resource, id).Error
}

func (s *GenericService) GetResourceByStringID(id string, resource interface{}) error {
	return s.DB.Where("id = ?", id).First(resource).Error
}

// DeleteResourceByID deletes a resource by its integer ID
func (s *GenericService) DeleteResourceByID(id int, resource interface{}) (int64, error) {
	result := s.DB.Delete(resource, id)
	return result.RowsAffected, result.Error
}

// DeleteResourceByStringID deletes a resource by its string ID
func (s *GenericService) DeleteResourceByStringID(id string, resource interface{}) (int64, error) {
	result := s.DB.Delete(resource, "id = ?", id)
	return result.RowsAffected, result.Error
}

// GetDepartmentsWithFaculty retrieves all departments with their faculty preloaded
func (s *GenericService) GetDepartmentsWithFaculty() ([]models.Department, error) {
	var departments []models.Department
	err := s.DB.Preload("Faculty").Find(&departments).Error
	return departments, err
} 
