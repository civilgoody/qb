package handlers

import (
	"fmt"
	"qb/pkg/database"
	"qb/pkg/models"
	"qb/pkg/utils"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

var validate = validator.New()

func Status(c *gin.Context) {
	utils.SuccessResponse(c, gin.H{"message": "Welcome to the Qb API"})
}

func CreateFaculty(c *gin.Context) {
	var faculty models.Faculty
	
	if err := c.ShouldBindJSON(&faculty); err != nil {
		utils.HandleValidationError(c, err)
		return
	}

	if err := validate.Struct(faculty); err != nil {
		utils.HandleValidationError(c, err)
		return
	}

	if err := database.DB.Create(&faculty).Error; err != nil {
		utils.HandleDatabaseError(c, err)
		return
	}

	utils.SuccessResponse(c, faculty)
}

func GetFaculties(c *gin.Context) {
	var faculties []models.Faculty
	
	if err := database.DB.Find(&faculties).Error; err != nil {
		utils.HandleError(c, utils.ErrDatabase)
		return
	}

	utils.SuccessResponse(c, faculties)
}

func DeleteFaculty(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		utils.HandleError(c, utils.NewValidationError("Invalid faculty ID format"))
		return
	}

	result := database.DB.Delete(&models.Faculty{}, id)
	if result.Error != nil {
		utils.HandleError(c, utils.ErrDatabase)
		return
	}

	if result.RowsAffected == 0 {
		utils.HandleError(c, utils.ErrNotFound)
		return
	}

	utils.SuccessResponse(c, gin.H{"message": fmt.Sprintf("Faculty with ID %d deleted successfully", id)})
}

func CreateLevel(c *gin.Context) {
	var level models.Level

	if err := c.ShouldBindJSON(&level); err != nil {
		utils.HandleValidationError(c, err)
		return
	}

	if err := validate.Struct(level); err != nil {
		utils.HandleValidationError(c, err)
		return
	}

	if err := database.DB.Create(&level).Error; err != nil {
		utils.HandleDatabaseError(c, err)
		return
	}

	utils.SuccessResponse(c, level)
}

func GetLevels(c *gin.Context) {
	var levels []models.Level
	
	if err := database.DB.Find(&levels).Error; err != nil {
		utils.HandleError(c, utils.ErrDatabase)
		return
	}

	utils.SuccessResponse(c, levels)
}

func DeleteLevel(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		utils.HandleError(c, utils.NewValidationError("Invalid level ID format"))
		return
	}

	result := database.DB.Delete(&models.Level{}, id)
	if result.Error != nil {
		utils.HandleError(c, utils.ErrDatabase)
		return
	}

	if result.RowsAffected == 0 {
		utils.HandleError(c, utils.ErrNotFound)
		return
	}

	utils.SuccessResponse(c, gin.H{"message": fmt.Sprintf("Level with ID %d deleted successfully", id)})
}
