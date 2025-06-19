package models

import (
	"time"
)

// Todo represents a single To-Do item in the database.
type Todo struct {
	// gorm.Model includes ID, CreatedAt, UpdatedAt, DeletedAt (soft delete).
	// We'll define our own ID to match the string type requirement,
	// but keeping the other GORM fields for convenience.
	ID          string         `gorm:"primaryKey;type:char(36);default:(uuid())" json:"id" example:"a1b2c3d4-e5f6-7890-1234-567890abcdef"` // Changed for MySQL UUID
	Title       string         `json:"title" binding:"required" validate:"required,min=3,max=100" example:"Buy groceries"`
	Description *string        `json:"description" validate:"omitempty,max=500" example:"Milk, eggs, bread and cheese"`
	Completed   bool           `gorm:"default:false" json:"completed" example:"false"`
	CreatedAt   time.Time      `gorm:"autoCreateTime" json:"created_at" example:"2023-01-01T12:00:00Z"`
	UpdatedAt   time.Time      `gorm:"autoUpdateTime" json:"updated_at" example:"2023-01-01T12:30:00Z"`
	DeletedAt   *time.Time     `json:"deleted_at,omitempty" example:"null"` // Simplified for Swagger parsing
} 
