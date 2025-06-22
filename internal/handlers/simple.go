package handlers

import (
	"qb/pkg/database"
	"qb/pkg/models"
	"qb/pkg/utils"

	"github.com/gin-gonic/gin"
)

func Status(c *gin.Context) {
	utils.SuccessResponse(c, gin.H{"message": "Welcome to the Qb API"})
}

func CreateFaculty(c *gin.Context) {
	var faculty models.Faculty
	
	if utils.HandleCreateResource(c, database.DB, utils.Validator, &faculty) {
		return
	}
}

func GetFaculties(c *gin.Context) {
	var faculties []models.Faculty
	
	if utils.HandleGetResources(c, database.DB, &faculties) {
		return
	}
}

func DeleteFaculty(c *gin.Context) {
	if utils.DeleteResourceByID(c, database.DB, "id", "Faculty", &models.Faculty{}) {
		return
	}
}

func CreateLevel(c *gin.Context) {
	var level models.Level

	if utils.HandleCreateResource(c, database.DB, utils.Validator, &level) {
		return
	}
}

func GetLevels(c *gin.Context) {
	var levels []models.Level
	
	if utils.HandleGetResources(c, database.DB, &levels) {
		return
	}
}

func DeleteLevel(c *gin.Context) {
	if utils.DeleteResourceByID(c, database.DB, "id", "Level", &models.Level{}) {
		return
	}
}

func CreateDepartment(c *gin.Context) {
	var department models.Department
	
	if utils.HandleCreateResource(c, database.DB, utils.Validator, &department) {
		return
	}
}

func GetDepartments(c *gin.Context) {
	var departments []models.Department
	
	if err := database.DB.Preload("Faculty").Find(&departments).Error; err != nil {
		utils.HandleError(c, utils.ErrDatabase)
		return
	}
	utils.SuccessResponse(c, departments)
}

func DeleteDepartment(c *gin.Context) {
	if utils.DeleteResourceByStringID(c, database.DB, "id", "Department", &models.Department{}) {
		return
	}
}

func CreateSession(c *gin.Context) {
	var session models.Session

	if utils.HandleCreateResource(c, database.DB, utils.Validator, &session) {
		return
	}
}

func GetSessions(c *gin.Context) {
	var sessions []models.Session

	if utils.HandleGetResources(c, database.DB, &sessions) {
		return
	}
}

func DeleteSession(c *gin.Context) {
	if utils.DeleteResourceByStringID(c, database.DB, "id", "Session", &models.Session{}) {
		return
	}
}
