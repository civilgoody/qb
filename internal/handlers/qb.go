package handlers

import (
	"fmt"
	"qb/pkg/database"
	"qb/pkg/models"
	"net/http"

	"github.com/gin-gonic/gin"
)

func Status(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "Welcome to the Qb API"})
}

func CreateFaculty(c *gin.Context) {
	var faculty models.Faculty
	fmt.Println(c.Request.Body)
	if err := c.ShouldBindJSON(&faculty); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	fmt.Println(faculty)

	database.DB.Create(&faculty)

	c.JSON(http.StatusCreated, faculty)
}

func GetFaculties(c *gin.Context) {
	var faculties []models.Faculty
	database.DB.Find(&faculties)

	c.JSON(http.StatusOK, faculties)
}
