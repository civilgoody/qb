package handlers

import (
	"fmt"
	"qb/internal/services"
	"qb/pkg/models"
	"strconv"

	"github.com/gin-gonic/gin"
)

func Status(c *gin.Context) {
	Res.Send(c, gin.H{"message": "Welcome to the Qb API"}, nil)
}

// Faculty handlers
func CreateFaculty(c *gin.Context) {
	var faculty models.Faculty
	if err := c.ShouldBindJSON(&faculty); err != nil {
		Res.Invalid(c, err)
		return
	}

	err := services.CreateResource(&faculty)
	Res.Created(c, faculty, err)
}

func GetFaculties(c *gin.Context) {
	var faculties []models.Faculty
	err := services.GetAllResources(&faculties)
	Res.Send(c, faculties, err)
}

func DeleteFaculty(c *gin.Context) {
	deleteResourceByIntID(c, &models.Faculty{}, "Faculty")
}

// Level handlers
func CreateLevel(c *gin.Context) {
	var level models.Level
	if err := c.ShouldBindJSON(&level); err != nil {
		Res.Invalid(c, err)
		return
	}

	err := services.CreateResource(&level)
	Res.Created(c, level, err)
}

func GetLevels(c *gin.Context) {
	var levels []models.Level
	err := services.GetAllResources(&levels)
	Res.Send(c, levels, err)
}

func DeleteLevel(c *gin.Context) {
	deleteResourceByIntID(c, &models.Level{}, "Level")
}

// Department handlers
func CreateDepartment(c *gin.Context) {
	var department models.Department
	if err := c.ShouldBindJSON(&department); err != nil {
		Res.Invalid(c, err)
		return
	}

	err := services.CreateResource(&department)
	Res.Created(c, department, err)
}

func GetDepartments(c *gin.Context) {
	departments, err := services.GetDepartmentsWithFaculty()
	Res.Send(c, departments, err)
}

func DeleteDepartment(c *gin.Context) {
	deleteResourceByStringID(c, &models.Department{}, "Department")
}

// Session handlers
func CreateSession(c *gin.Context) {
	var session models.Session
	if err := c.ShouldBindJSON(&session); err != nil {
		Res.Invalid(c, err)
		return
	}

	err := services.CreateResource(&session)
	Res.Created(c, session, err)
}

func GetSessions(c *gin.Context) {
	var sessions []models.Session
	err := services.GetAllResources(&sessions)
	Res.Send(c, sessions, err)
}

func DeleteSession(c *gin.Context) {
	deleteResourceByStringID(c, &models.Session{}, "Session")
}

// Request handlers
func GetRequests(c *gin.Context) {
	var requests []models.TemporaryUpload
	err := services.GetAllResources(&requests)
	Res.Send(c, requests, err)
}

// Helper functions

// deleteResourceByIntID handles deletion for resources with integer IDs
func deleteResourceByIntID(c *gin.Context, resource interface{}, resourceName string) {
	id, err := parseIntID(c, "id")
	if err != nil {
		Res.Invalid(c, err)
		return
	}

	err = services.DeleteResource(id, resource)
	Res.Send(c, nil, err, resourceName+" deleted successfully")
}

// deleteResourceByStringID handles deletion for resources with string IDs  
func deleteResourceByStringID(c *gin.Context, resource interface{}, resourceName string) {
	id := c.Param("id")
	err := services.DeleteResource(id, resource)
	Res.Send(c, nil, err, resourceName+" deleted successfully")
}

// parseIntID parses integer ID from gin context parameter
func parseIntID(c *gin.Context, paramName string) (int, error) {
	idStr := c.Param(paramName)
	id, err := strconv.Atoi(idStr)
	if err != nil {
		return 0, fmt.Errorf(paramName + " must be a valid integer")
	}
	return id, nil
}
