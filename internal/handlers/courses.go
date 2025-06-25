package handlers

import (
	"qb/internal/services"
	"qb/pkg/database"
	"qb/pkg/models"
	"qb/pkg/utils"
	"strconv"

	"github.com/gin-gonic/gin"
)

var courseService *services.CourseService

// InitCourseService initializes the course service
func InitCourseService() {
	courseService = services.NewCourseService(database.DB, utils.Validator)
}

// CreateCourse handles course creation with business logic
func CreateCourse(c *gin.Context) {
	var input models.CreateCourseDTO
	
	if err := c.ShouldBindJSON(&input); err != nil {
		utils.HandleValidationError(c, err)
		return
	}

	course, err := courseService.CreateCourse(input)
	if err != nil {
		utils.HandleDatabaseError(c, err)
		return
	}

	utils.SuccessResponse(c, course)
}

// GetCourses retrieves all courses with their relationships
func GetCourses(c *gin.Context) {
	courses, err := courseService.GetAllCourses()
	if err != nil {
		utils.HandleDatabaseError(c, err)
		return
	}

	utils.SuccessResponse(c, courses)
}

// FilterCourses handles course filtering by department, level, and semester
func FilterCourses(c *gin.Context) {
	dept, level, semester, ok := parseFilterParams(c)
	if !ok {
		return
	}

	courses, err := courseService.FilterCourses(dept, level, semester)
	if err != nil {
		utils.HandleDatabaseError(c, err)
		return
	}

	utils.SuccessResponse(c, courses)
}

// DeleteCourse handles course deletion
func DeleteCourse(c *gin.Context) {
	DeleteByStringID(c, &models.Course{}, "Course")
}

// Helper function to parse and validate course filter parameters
func parseFilterParams(c *gin.Context) (dept string, level int, semester int, ok bool) {
	dept = c.Param("dept")
	
	// Parse level
	levelStr := c.Param("level")
	var err error
	level, err = strconv.Atoi(levelStr)
	if err != nil {
		utils.HandleError(c, utils.NewValidationError("level must be a valid integer"))
		return "", 0, 0, false
	}
	
	// Parse semester
	semesterStr := c.Param("semester")
	semester, err = strconv.Atoi(semesterStr)
	if err != nil {
		utils.HandleError(c, utils.NewValidationError("semester must be a valid integer"))
		return "", 0, 0, false
	}
	
	// Validate level (100, 200, 300, 400, 500)
	if level < 100 || level > 500 || level%100 != 0 {
		utils.HandleError(c, utils.NewValidationError("Level must be one of: 100, 200, 300, 400, 500"))
		return "", 0, 0, false
	}
	
	// Validate semester (1 or 2)
	if semester < 1 || semester > 2 {
		utils.HandleError(c, utils.NewValidationError("Semester must be 1 or 2"))
		return "", 0, 0, false
	}
	
	return dept, level, semester, true
}
