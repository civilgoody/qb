package models

import (
	"fmt"
	"strconv"
	"strings"
)

type CreateCourseDTO struct {
	Code        string        `json:"code" binding:"required" validate:"required,len=6"`
	Units       int           `json:"units" binding:"required" validate:"required,min=1,max=10"`
	Title       string        `json:"title" binding:"required" validate:"required"`
	Description *string       `json:"description,omitempty"`
	Status      *CourseStatus `json:"status,omitempty"`
}

// ParseCourseCode parses a 6-character course code (e.g., "CEG543")
// Returns: departmentCode, level, semester, finalDigit, error
func (dto *CreateCourseDTO) ParseCourseCode() (string, int, int, error) {
	code := strings.ToUpper(dto.Code)
	
	if len(code) != 6 {
		return "", 0, 0, fmt.Errorf("course code must be exactly 6 characters")
	}

	// Extract department code (first 3 characters)
	departmentCode := code[:3]

	// Extract level (4th character)
	levelChar := code[3:4]
	levelDigit, err := strconv.Atoi(levelChar)
	if err != nil || levelDigit < 1 || levelDigit > 5 {
		return "", 0, 0, fmt.Errorf("level must be a digit between 1-5")
	}
	level := levelDigit * 100 // Convert 1->100, 2->200, etc.

	// Extract semester indicator (5th character)
	semesterChar := code[4:5]
	semesterDigit, err := strconv.Atoi(semesterChar)
	if err != nil {
		return "", 0, 0, fmt.Errorf("semester indicator must be a digit")
	}
	
	var semester int
	if semesterDigit%2 == 1 { // Odd = 1st semester
		semester = 1
	} else { // Even = 2nd semester
		semester = 2
	}

	// Extract final digit (6th character) - user's choice
	// finalDigit := 
	if _, err := strconv.Atoi(code[5:6]); err != nil {
		return "", 0, 0, fmt.Errorf("final character must be a digit")
	}

	return departmentCode, level, semester, nil
}

// Auth DTOs
type LoginDTO struct {
	Email    string `json:"email" binding:"required,email" validate:"required,email"`
	Password string `json:"password" binding:"required,min=6" validate:"required,min=6"`
}

type RegisterDTO struct {
	FirstName    string  `json:"firstName" binding:"required" validate:"required"`
	LastName     *string `json:"lastName,omitempty"`
	Email        string  `json:"email" binding:"required,email" validate:"required,email"`
	Password     string  `json:"password" binding:"required,min=6" validate:"required,min=6"`
	DepartmentID *string `json:"departmentId,omitempty"`
	LevelID      *int    `json:"levelId,omitempty"`
	Semester     *int    `json:"semester,omitempty"`
}

type AuthResponse struct {
	Token        string `json:"token"`
	RefreshToken string `json:"refreshToken"`
	User         User   `json:"user"`
}

type RefreshTokenDTO struct {
	RefreshToken string `json:"refreshToken" binding:"required" validate:"required"`
}
