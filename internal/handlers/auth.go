package handlers

import (
	"qb/internal/services"
	"qb/pkg/database"
	"qb/pkg/models"
	"qb/pkg/utils"

	"github.com/gin-gonic/gin"
)

var authService *services.AuthService

// InitAuthService initializes the auth service
func InitAuthService() {
	authService = services.NewAuthService(database.DB, utils.Validator)
}

// Register handles user registration
func Register(c *gin.Context) {
	var input models.RegisterDTO
	if err := c.ShouldBindJSON(&input); err != nil {
		Res.Invalid(c, err)
		return
	}

	response, err := authService.Register(input)
	Res.Send(c, response, err)
}

// Login handles user authentication
func Login(c *gin.Context) {
	var input models.LoginDTO
	if err := c.ShouldBindJSON(&input); err != nil {
		Res.Invalid(c, err)
		return
	}

	response, err := authService.Login(input)
	Res.Send(c, response, err)
}

// GetProfile handles getting the current user's profile
func GetProfile(c *gin.Context) {
	userID, err := Auth.GetCurrentUserID(c)
	if err != nil {
		Res.Send(c, nil, err)
		return
	}

	user, err := authService.GetProfile(userID)
	Res.Send(c, user, err)
}

// RefreshToken handles token refresh
func RefreshToken(c *gin.Context) {
	var input models.RefreshTokenDTO
	if err := c.ShouldBindJSON(&input); err != nil {
		Res.Invalid(c, err)
		return
	}

	newToken, err := authService.RefreshToken(input)
	if err != nil {
		Res.Send(c, nil, err)
		return
	}

	response := map[string]string{
		"token": newToken,
	}

	Res.Send(c, response, nil)
} 
 