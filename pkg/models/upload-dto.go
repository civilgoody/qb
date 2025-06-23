package models

import "mime/multipart"

// UploadImagesDTO represents the input for image upload requests
type UploadImagesDTO struct {
	ImageFiles []*multipart.FileHeader `form:"imageFiles" binding:"required,max=5" validate:"required,max=5"`
}

// UploadResult represents the result of uploading a single image
type UploadResult struct {
	OriginalFilename string `json:"originalFilename"`
	PublicID         string `json:"publicId,omitempty"`
	Error            string `json:"error,omitempty"`
}

// UploadResponse represents the response from the image upload endpoint
type UploadResponse struct {
	RequestID string         `json:"requestId"`
	Results   []UploadResult `json:"results"`
	Success   bool           `json:"success"`
}

// CreateQuestionDTO represents the enhanced question creation input with image handling
type CreateQuestionDTO struct {
	CourseID      string       `json:"courseId" binding:"required" validate:"required,len=6"`
	SessionID     string       `json:"sessionId" binding:"required" validate:"required"`
	Type          QuestionType `json:"type" binding:"required" validate:"required,oneof=TEST EXAM"`
	Lecturer      *string      `json:"lecturer,omitempty"`
	TimeAllowed   *int         `json:"timeAllowed,omitempty" validate:"omitempty,min=1,max=600"`
	DocLink       *string      `json:"docLink,omitempty" validate:"omitempty,url"`
	Tips          *string      `json:"tips,omitempty"`
	
	// Simplified image handling - just pass the entire upload response
	UploadResults *UploadResponse `json:"uploadResults,omitempty"`
}

// QuestionResponse represents the response after creating a question
type QuestionResponse struct {
	ID               string   `json:"id"`
	CourseID         string   `json:"courseId"`
	SessionID        string   `json:"sessionId"`
	Type             string   `json:"type"`
	ImageCount       int      `json:"imageCount"`
	ProcessingStatus string   `json:"processingStatus"`
	Approved         bool     `json:"approved"`
	CreatedAt        string   `json:"createdAt"`
	Message          string   `json:"message,omitempty"`
} 
