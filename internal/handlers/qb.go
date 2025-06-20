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
