package services

import (
	"qb/pkg/models"
	"qb/pkg/utils"

	"gorm.io/gorm"
)

// Register handles user registration business logic
func Register(input models.RegisterDTO) (*models.AuthResponse, error) {
	// Validate the DTO
	if err := valS.Struct(input); err != nil {
		return nil, errS.Invalid(err)
	}

	// Check if user already exists
	var existingUser models.User
	if err := db.Where("email = ?", input.Email).First(&existingUser).Error; err == nil {
		return nil, errS.Invalid("User with this email already exists")
	} else if err != gorm.ErrRecordNotFound {
		return nil, errS.Db(err)
	}

	// Hash the password
	hashedPassword, err := utils.HashPassword(input.Password)
	if err != nil {
		return nil, errS.Invalid("Failed to process password")
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

	if err := db.Create(&user).Error; err != nil {
		return nil, errS.Db(err)
	}

	// Generate tokens
	token, err := utils.GenerateToken(user.ID, user.Email, string(user.Role))
	if err != nil {
		return nil, errS.Invalid("Failed to generate access token")
	}

	refreshToken, err := utils.GenerateRefreshToken(user.ID)
	if err != nil {
		return nil, errS.Invalid("Failed to generate refresh token")
	}

	// Remove password from response
	user.Password = nil

	response := &models.AuthResponse{
		Token:        token,
		RefreshToken: refreshToken,
		User:         user,
	}

	return response, nil
}

// Login handles user authentication business logic
func Login(input models.LoginDTO) (*models.AuthResponse, error) {
	// Validate the DTO
	if err := valS.Struct(input); err != nil {
		return nil, errS.Invalid(err)
	}

	// Find user by email
	var user models.User
	if err := db.Where("email = ? AND is_active = ?", input.Email, true).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, models.ErrUnauthorized
		}
		return nil, errS.Db(err)
	}

	// Check password
	if user.Password == nil || !utils.CheckPassword(*user.Password, input.Password) {
		return nil, models.ErrBadLogin
	}

	// Generate tokens
	token, err := utils.GenerateToken(user.ID, user.Email, string(user.Role))
	if err != nil {
		return nil, errS.Invalid("Failed to generate access token")
	}

	refreshToken, err := utils.GenerateRefreshToken(user.ID)
	if err != nil {
		return nil, errS.Invalid("Failed to generate refresh token")
	}

	// Remove password from response
	user.Password = nil

	response := &models.AuthResponse{
		Token:        token,
		RefreshToken: refreshToken,
		User:         user,
	}

	return response, nil
}

// GetProfile retrieves user profile by ID
func GetProfile(userID string) (*models.User, error) {
	var user models.User
	if err := db.Preload("Department").Preload("Level").
		Where("id = ? AND is_active = ?", userID, true).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, models.ErrNotFound
		}
		return nil, errS.Db(err)
	}

	// Remove password from response
	user.Password = nil

	return &user, nil
}

// RefreshToken handles token refresh business logic
func RefreshToken(input models.RefreshTokenDTO) (string, error) {
	// Validate the DTO
	if err := valS.Struct(input); err != nil {
		return "", errS.Invalid(err)
	}

	// Validate refresh token
	userID, err := utils.ValidateRefreshToken(input.RefreshToken)
	if err != nil {
		return "", models.ErrUnauthorized
	}

	// Get user details
	var user models.User
	if err := db.Where("id = ? AND is_active = ?", userID, true).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return "", models.ErrUnauthorized
		}
		return "", errS.Db(err)
	}

	// Generate new access token
	newToken, err := utils.GenerateToken(user.ID, user.Email, string(user.Role))
	if err != nil {
		return "", errS.Invalid("Failed to generate token")
	}

	return newToken, nil
} 
