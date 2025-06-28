package services

import (
	"qb/pkg/models"

	"gorm.io/gorm"
)

// CreateResource creates a new resource in the database
func CreateResource(resource interface{}) error {
	if err := valS.Struct(resource); err != nil {
		return errS.Invalid(err)
	}

	if err := db.Create(resource).Error; err != nil {
		return errS.Db(err)
	}

	return nil
}

// GetAllResources retrieves all resources of a given type
func GetAllResources(resources interface{}) error {
	if err := db.Find(resources).Error; err != nil {
		return errS.Db(err)
	}
	return nil
}

// GetResourceByID gets a resource by integer ID
func GetResourceByID(id int, resource interface{}) error {
	if err := db.First(resource, id).Error; err != nil {
		return errS.Db(err)
	}
	return nil
}

// GetResourceByStringID gets a resource by string ID
func GetResourceByStringID(id string, resource interface{}) error {
	if err := db.Where("id = ?", id).First(resource).Error; err != nil {
		return errS.Db(err)
	}
	return nil
}

// DeleteResource deletes a resource by ID (handles both int and string IDs)
func DeleteResource(id interface{}, resource interface{}) error {
	var result *gorm.DB
		
	switch v := id.(type) {
	case int:
		result = db.Delete(resource, v)
	case string:
		result = db.Delete(resource, "id = ?", v)
	default:
		return errS.Invalid("ID must be either int or string")
	}
	
	if result.Error != nil {
		return errS.Db(result.Error)
	}
	
	if result.RowsAffected == 0 {
		return models.ErrNotFound
	}
	
	return nil
}

// GetDepartmentsWithFaculty retrieves all departments with their faculty preloaded
func GetDepartmentsWithFaculty() ([]models.Department, error) {
	var departments []models.Department
	if err := db.Preload("Faculty").Find(&departments).Error; err != nil {
		return nil, errS.Db(err)
	}
	return departments, nil
} 
