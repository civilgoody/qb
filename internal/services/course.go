package services

import (
	"qb/pkg/models"
)

// CreateCourse creates a new course with parsed course code validation
func CreateCourse(input models.CreateCourseDTO) (*models.Course, error) {
	// Validate the DTO
	if err := valS.Struct(input); err != nil {
		return nil, errS.Invalid(err)
	}

	// Parse and validate course code
	departmentCode, level, semester, err := input.ParseCourseCode()
	if err != nil {
		return nil, errS.Invalid(err)
	}

	// Create the course entity
	course := models.Course{
		ID:          input.Code,
		Units:       input.Units,
		Title:       input.Title,
		LevelID:     level,
		Semester:    semester,
		Description: input.Description,
		Status:      input.Status,
	}

	// Verify the department exists and level exists
	var department models.Department
	if err := db.Where("id = ?", departmentCode).First(&department).Error; err != nil {
		return nil, err
	}

	var levelModel models.Level
	if err := db.Where("id = ?", level).First(&levelModel).Error; err != nil {
		return nil, err
	}

	// Create the course
	if err := db.Create(&course).Error; err != nil {
		return nil, err
	}

	// Associate with department (many-to-many relationship)
	if err := db.Model(&course).Association("Departments").Append(&department); err != nil {
		return nil, err
	}

	return &course, nil
}

// GetAllCourses retrieves all courses with their relationships
func GetAllCourses() ([]models.Course, error) {
	var courses []models.Course
	if err := db.Find(&courses).Error; err != nil {
		return nil, err
	}
	return courses, nil
}

// FilterCourses filters courses by department, level, and semester
func FilterCourses(filter models.CourseFilterDTO) ([]models.Course, error) {
	var courses []models.Course
	
	query := db.
		Where("level_id = ? AND semester = ?", filter.Level, filter.Semester).
		Joins("JOIN department_courses ON courses.id = department_courses.course_id").
		Joins("JOIN departments ON department_courses.department_id = departments.id").
		Where("departments.id = ?", filter.Dept)

	if err := query.Find(&courses).Error; err != nil {
		return nil, err
	}

	return courses, nil
}

// DeleteCourse deletes a course by its ID
func DeleteCourse(id string) (int64, error) {
	result := db.Delete(&models.Course{}, "id = ?", id)
	if result.Error != nil {
		return 0, result.Error
	}
	
	if result.RowsAffected == 0 {
		return 0, models.ErrNotFound
	}
	
	return result.RowsAffected, nil
} 
