package handlers

import (
	"qb/internal/services"
	"qb/pkg/models"

	"github.com/gin-gonic/gin"
)

// Register handles user registration
func Register(c *gin.Context) {
	var input models.RegisterDTO
	if err := c.ShouldBindJSON(&input); err != nil {
		Res.Invalid(c, err)
		return
	}

	response, err := services.Register(input)
	Res.Created(c, response, err)
}

// Login handles user authentication
func Login(c *gin.Context) {
	var input models.LoginDTO
	if err := c.ShouldBindJSON(&input); err != nil {
		Res.Invalid(c, err)
		return
	}

	response, err := services.Login(input)
	Res.Send(c, response, err, "Login successful")
}

// GetProfile retrieves the authenticated user's profile
func GetProfile(c *gin.Context) {
	userID := c.GetString("userID")

	user, err := services.GetProfile(userID)
	Res.Send(c, user, err, "Profile retrieved successfully")
}

// RefreshToken handles token refresh
func RefreshToken(c *gin.Context) {
	var input models.RefreshTokenDTO
	if err := c.ShouldBindJSON(&input); err != nil {
		Res.Invalid(c, err)
		return
	}

	newToken, err := services.RefreshToken(input)
	response := gin.H{"token": newToken}
	Res.Send(c, response, err, "Token refreshed successfully")
} 
 