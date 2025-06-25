package handlers

import (
	"qb/internal/services"
	"qb/pkg/database"
	"qb/pkg/models"
	"qb/pkg/utils"
	"strconv"

	"github.com/gin-gonic/gin"
)

var s *services.GenericService

// InitGenericService initializes the Generic service with database connection
func InitGenericService() {
	s = services.NewGenericService(database.DB, utils.Validator)
}

func Status(c *gin.Context) {
	utils.SuccessResponse(c, gin.H{"message": "Welcome to the Qb API"})
}

// Faculty handlers
func CreateFaculty(c *gin.Context) {
	var faculty models.Faculty
	
	if err := c.ShouldBindJSON(&faculty); err != nil {
		utils.HandleValidationError(c, err)
		return
	}

	if err := s.CreateResource(&faculty); err != nil {
		utils.HandleDatabaseError(c, err)
		return
	}

	utils.SuccessResponse(c, faculty)
}

func GetFaculties(c *gin.Context) {
	var faculties []models.Faculty
	
	if err := s.GetAllResources(&faculties); err != nil {
		utils.HandleDatabaseError(c, err)
		return
	}

	utils.SuccessResponse(c, faculties)
}

func DeleteFaculty(c *gin.Context) {
	deleteByID(c, &models.Faculty{}, "Faculty")
}

// Level handlers
func CreateLevel(c *gin.Context) {
	var level models.Level

	if err := c.ShouldBindJSON(&level); err != nil {
		utils.HandleValidationError(c, err)
		return
	}

	if err := s.CreateResource(&level); err != nil {
		utils.HandleDatabaseError(c, err)
		return
	}

	utils.SuccessResponse(c, level)
}

func GetLevels(c *gin.Context) {
	var levels []models.Level
	
	if err := s.GetAllResources(&levels); err != nil {
		utils.HandleDatabaseError(c, err)
		return
	}

	utils.SuccessResponse(c, levels)
}

func DeleteLevel(c *gin.Context) {
	deleteByID(c, &models.Level{}, "Level")
}

// Department handlers
func CreateDepartment(c *gin.Context) {
	var department models.Department
	
	if err := c.ShouldBindJSON(&department); err != nil {
		utils.HandleValidationError(c, err)
		return
	}

	if err := s.CreateResource(&department); err != nil {
		utils.HandleDatabaseError(c, err)
		return
	}

	utils.SuccessResponse(c, department)
}

func GetDepartments(c *gin.Context) {
	departments, err := s.GetDepartmentsWithFaculty()
	if err != nil {
		utils.HandleDatabaseError(c, err)
		return
	}

	utils.SuccessResponse(c, departments)
}

func DeleteDepartment(c *gin.Context) {
	deleteByStringID(c, &models.Department{}, "Department")
}

// Session handlers
func CreateSession(c *gin.Context) {
	var session models.Session

	if err := c.ShouldBindJSON(&session); err != nil {
		utils.HandleValidationError(c, err)
		return
	}

	if err := s.CreateResource(&session); err != nil {
		utils.HandleDatabaseError(c, err)
		return
	}

	utils.SuccessResponse(c, session)
}

func GetSessions(c *gin.Context) {
	var sessions []models.Session

	if err := s.GetAllResources(&sessions); err != nil {
		utils.HandleDatabaseError(c, err)
		return
	}

	utils.SuccessResponse(c, sessions)
}

func DeleteSession(c *gin.Context) {
	deleteByStringID(c, &models.Session{}, "Session")
}


// Request handlers
func GetRequests(c *gin.Context) {
	var requests []models.TemporaryUpload

	if err := s.GetAllResources(&requests); err != nil {
		utils.HandleDatabaseError(c, err)
		return
	}

	utils.SuccessResponse(c, requests)
}

/*
**	Helper functions
*/

// Helper function to parse integer ID from gin context
func parseIntIDFromParam(c *gin.Context, paramName string) (int, bool) {
	idStr := c.Param(paramName)
	id, err := strconv.Atoi(idStr)
	if err != nil {
		utils.HandleError(c, utils.NewValidationError(paramName+" must be a valid integer"))
		return 0, false
	}
	return id, true
}

func deleteByID(c *gin.Context, resource interface{}, resourceType string) {
	id, ok := parseIntIDFromParam(c, "id")
	if !ok {
		return
	}

	rowsAffected, err := s.DeleteResourceByID(id, resource)
	if err != nil {
		utils.HandleDatabaseError(c, err)
		return
	}

	if rowsAffected == 0 {
		utils.HandleError(c, utils.ErrNotFound)
		return
	}

	utils.SuccessResponse(c, gin.H{"message": resourceType + " deleted successfully"})
}

func deleteByStringID(c *gin.Context, resource interface{}, resourceType string) {
	id := c.Param("id")

	rowsAffected, err := s.DeleteResourceByStringID(id, resource)
	if err != nil {
		utils.HandleDatabaseError(c, err)
		return
	}

	if rowsAffected == 0 {
		utils.HandleError(c, utils.ErrNotFound)
		return
	}

	utils.SuccessResponse(c, gin.H{"message": resourceType + " deleted successfully"})
}
