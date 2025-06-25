package handlers

import (
	"net/http"
	"qb/pkg/database"
	"qb/pkg/models"
	"qb/pkg/utils"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// Register handles user registration
func Register(c *gin.Context) {
	var input models.RegisterDTO
	if utils.BindAndValidate(c, &input) {
		return
	}

	// Check if user already exists
	var existingUser models.User
	if err := database.DB.Where("email = ?", input.Email).First(&existingUser).Error; err == nil {
		utils.HandleError(c, utils.NewValidationError("User with this email already exists"))
		return
	} else if err != gorm.ErrRecordNotFound {
		utils.HandleDatabaseError(c, err)
		return
	}

	// Hash the password
	hashedPassword, err := utils.HashPassword(input.Password)
	if err != nil {
		utils.HandleError(c, utils.NewValidationError("Failed to process password"))
		return
	}

	// Create user
	user := models.User{
		FirstName:    input.FirstName,
		LastName:     input.LastName,
		Email:        input.Email,
		Password:     &hashedPassword,
		DepartmentID: input.DepartmentID,
		LevelID:      input.LevelID,
		Semester:     input.Semester,
		Role:         models.RoleMember, // Default role
		IsActive:     true,
	}

	if err := database.DB.Create(&user).Error; err != nil {
		utils.HandleDatabaseError(c, err)
		return
	}

	// Generate JWT token
	token, err := utils.GenerateToken(user.ID, user.Email, string(user.Role))
	if err != nil {
		utils.HandleError(c, utils.NewValidationError("Failed to generate access token"))
		return
	}

	refreshToken, err := utils.GenerateRefreshToken(user.ID)
	if err != nil {
		utils.HandleError(c, utils.NewValidationError("Failed to generate refresh token"))
		return
	}

	// Remove password from response
	user.Password = nil

	response := models.AuthResponse{
		Token:        token,
		RefreshToken: refreshToken,
		User:         user,
	}

	utils.SuccessResponse(c, response)
}

// Login handles user authentication
func Login(c *gin.Context) {
	var input models.LoginDTO
	if utils.BindAndValidate(c, &input) {
		return
	}

	// Find user by email
	var user models.User
	if err := database.DB.Where("email = ? AND is_active = ?", input.Email, true).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			utils.ErrorResponse(c, http.StatusUnauthorized, utils.ErrUnauthorized, "Invalid email or password")
			return
		}
		utils.HandleDatabaseError(c, err)
		return
	}

	// Check password
	if user.Password == nil || !utils.CheckPassword(*user.Password, input.Password) {
		utils.ErrorResponse(c, http.StatusUnauthorized, utils.ErrUnauthorized, "Invalid email or password")
		return
	}

	// Generate JWT token
	token, err := utils.GenerateToken(user.ID, user.Email, string(user.Role))
	if err != nil {
		utils.HandleError(c, utils.NewValidationError("Failed to generate access token"))
		return
	}

	refreshToken, err := utils.GenerateRefreshToken(user.ID)
	if err != nil {
		utils.HandleError(c, utils.NewValidationError("Failed to generate refresh token"))
		return
	}

	// Remove password from response
	user.Password = nil

	response := models.AuthResponse{
		Token:        token,
		RefreshToken: refreshToken,
		User:         user,
	}

	utils.SuccessResponse(c, response)
}

// GetProfile handles getting the current user's profile
func GetProfile(c *gin.Context) {
	userID, exists := utils.GetCurrentUserID(c)
	if !exists {
		utils.ErrorResponse(c, http.StatusUnauthorized, utils.ErrUnauthorized, "User not found in context")
		return
	}

	var user models.User
	if err := database.DB.Preload("Department").Preload("Level").
		Where("id = ? AND is_active = ?", userID, true).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			utils.ErrorResponse(c, http.StatusNotFound, utils.ErrNotFound, "User not found")
			return
		}
		utils.HandleDatabaseError(c, err)
		return
	}

	// Remove password from response
	user.Password = nil

	utils.SuccessResponse(c, user)
}

// RefreshToken handles token refresh
func RefreshToken(c *gin.Context) {
	var input models.RefreshTokenDTO
	if utils.BindAndValidate(c, &input) {
		return
	}

	// Validate refresh token
	userID, err := utils.ValidateRefreshToken(input.RefreshToken)
	if err != nil {
		utils.ErrorResponse(c, http.StatusUnauthorized, utils.ErrUnauthorized, "Invalid refresh token")
		return
	}

	// Get user details
	var user models.User
	if err := database.DB.Where("id = ? AND is_active = ?", userID, true).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			utils.ErrorResponse(c, http.StatusUnauthorized, utils.ErrUnauthorized, "User not found")
			return
		}
		utils.HandleDatabaseError(c, err)
		return
	}

	// Generate new access token
	newToken, err := utils.GenerateToken(user.ID, user.Email, string(user.Role))
	if err != nil {
		utils.HandleError(c, utils.NewValidationError("Failed to generate token"))
		return
	}

	response := map[string]string{
		"token": newToken,
	}

	utils.SuccessResponse(c, response)
} 
