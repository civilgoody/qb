package services

import (
	"qb/pkg/models"
	"qb/pkg/utils"

	"github.com/go-playground/validator/v10"
	"gorm.io/gorm"
)

// AuthService handles authentication-specific business logic
type AuthService struct {
	db       *gorm.DB
	validate *validator.Validate
	err      *ErrorService
}

// NewAuthService creates a new auth service instance
func NewAuthService(db *gorm.DB, validate *validator.Validate) *AuthService {
	return &AuthService{
		db:       db,
		validate: validate,
		err:      NewErrorService(),
	}
}

// Register handles user registration business logic
func (s *AuthService) Register(input models.RegisterDTO) (*models.AuthResponse, error) {
	// Validate the DTO
	if err := s.validate.Struct(input); err != nil {
		return nil, s.err.Invalid(err)
	}

	// Check if user already exists
	var existingUser models.User
	if err := s.db.Where("email = ?", input.Email).First(&existingUser).Error; err == nil {
		return nil, s.err.Invalid("User with this email already exists")
	} else if err != gorm.ErrRecordNotFound {
		return nil, s.err.Db(err)
	}

	// Hash the password
	hashedPassword, err := utils.HashPassword(input.Password)
	if err != nil {
		return nil, s.err.Invalid("Failed to process password")
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

	if err := s.db.Create(&user).Error; err != nil {
		return nil, s.err.Db(err)
	}

	// Generate tokens
	token, err := utils.GenerateToken(user.ID, user.Email, string(user.Role))
	if err != nil {
		return nil, s.err.Invalid("Failed to generate access token")
	}

	refreshToken, err := utils.GenerateRefreshToken(user.ID)
	if err != nil {
		return nil, s.err.Invalid("Failed to generate refresh token")
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
func (s *AuthService) Login(input models.LoginDTO) (*models.AuthResponse, error) {
	// Validate the DTO
	if err := s.validate.Struct(input); err != nil {
		return nil, s.err.Invalid(err)
	}

	// Find user by email
	var user models.User
	if err := s.db.Where("email = ? AND is_active = ?", input.Email, true).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, models.ErrUnauthorized
		}
		return nil, s.err.Db(err)
	}

	// Check password
	if user.Password == nil || !utils.CheckPassword(*user.Password, input.Password) {
		return nil, models.ErrBadLogin
	}

	// Generate tokens
	token, err := utils.GenerateToken(user.ID, user.Email, string(user.Role))
	if err != nil {
		return nil, s.err.Invalid("Failed to generate access token")
	}

	refreshToken, err := utils.GenerateRefreshToken(user.ID)
	if err != nil {
		return nil, s.err.Invalid("Failed to generate refresh token")
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
func (s *AuthService) GetProfile(userID string) (*models.User, error) {
	var user models.User
	if err := s.db.Preload("Department").Preload("Level").
		Where("id = ? AND is_active = ?", userID, true).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, models.ErrNotFound
		}
		return nil, s.err.Db(err)
	}

	// Remove password from response
	user.Password = nil

	return &user, nil
}

// RefreshToken handles token refresh business logic
func (s *AuthService) RefreshToken(input models.RefreshTokenDTO) (string, error) {
	// Validate the DTO
	if err := s.validate.Struct(input); err != nil {
		return "", s.err.Invalid(err)
	}

	// Validate refresh token
	userID, err := utils.ValidateRefreshToken(input.RefreshToken)
	if err != nil {
		return "", models.ErrUnauthorized
	}

	// Get user details
	var user models.User
	if err := s.db.Where("id = ? AND is_active = ?", userID, true).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return "", models.ErrUnauthorized
		}
		return "", s.err.Db(err)
	}

	// Generate new access token
	newToken, err := utils.GenerateToken(user.ID, user.Email, string(user.Role))
	if err != nil {
		return "", s.err.Invalid("Failed to generate token")
	}

	return newToken, nil
} 
