package handlers

import (
	"qb/internal/services"
	"qb/pkg/models"

	"github.com/gin-gonic/gin"
)

// CreateCourse handles course creation
func CreateCourse(c *gin.Context) {
	var input models.CreateCourseDTO
	if err := c.ShouldBindJSON(&input); err != nil {
		Res.Invalid(c, err)
		return
	}

	course, err := services.CreateCourse(input)
	Res.Created(c, course, err)
}

// GetAllCourses handles retrieving all courses
func GetAllCourses(c *gin.Context) {
	courses, err := services.GetAllCourses()
	Res.Send(c, courses, err)
}

// FilterCourses handles filtering courses by department, level, and semester
func FilterCourses(c *gin.Context) {
	var filter models.CourseFilterDTO
	if err := c.ShouldBindJSON(&filter); err != nil {
		Res.Invalid(c, err)
		return
	}

	courses, err := services.FilterCourses(filter)
	Res.Send(c, courses, err)
}

// DeleteCourse handles course deletion
func DeleteCourse(c *gin.Context) {
	id := c.Param("id")
	_, err := services.DeleteCourse(id)
	Res.Send(c, gin.H{"deleted": true}, err, "Course deleted successfully")
}

