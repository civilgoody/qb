package handlers

import (
	"qb/internal/services"
	"qb/pkg/database"
	"qb/pkg/models"
	"qb/pkg/utils"

	"github.com/gin-gonic/gin"
)

var courseService *services.CourseService

// InitCourseService initializes the course service
func InitCourseService() {
	courseService = services.NewCourseService(database.DB, utils.Validator)
}

// CreateCourse handles course creation
func CreateCourse(c *gin.Context) {
	var input models.CreateCourseDTO

	course, err := courseService.CreateCourse(input)
	Res.Created(c, course, err)
}

// GetCourses retrieves all courses
func GetCourses(c *gin.Context) {
	courses, err := courseService.GetAllCourses()
	Res.Send(c, courses, err)
}

// FilterCourses handles course filtering by department, level, and semester
func FilterCourses(c *gin.Context) {
	var filter models.CourseFilterDTO
	if err := c.ShouldBindUri(&filter); err != nil {
		Res.Invalid(c, err)
		return
	}

	courses, err := courseService.FilterCourses(filter)
	Res.Send(c, courses, err)
}

// DeleteCourse handles course deletion
func DeleteCourse(c *gin.Context) {
	id := c.Param("id")
	
	_, err := courseService.DeleteCourse(id)
	Res.Send(c, nil, err, "Course deleted successfully")
}

