package handlers

import (
	"fmt"
	"qb/pkg/database"
	"qb/pkg/models"
	"qb/pkg/utils"
	"strconv"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// CreateQuestion handles question creation with image finalization
func CreateQuestion(c *gin.Context) {
	// Bind and validate the DTO
	var input models.CreateQuestionDTO
	if utils.BindAndValidate(c, &input) {
		return
	}

	// Validate that the course exists
	var course models.Course
	if err := database.DB.Where("id = ?", input.CourseID).First(&course).Error; err != nil {
		utils.HandleDatabaseErrorWithContext(c, err, "Course")
		return
	}

	// Validate that the session exists
	var session models.Session
	if err := database.DB.Where("id = ?", input.SessionID).First(&session).Error; err != nil {
		utils.HandleDatabaseErrorWithContext(c, err, "Session")
		return
	}

	// Extract and validate upload results if provided
	var tempPublicIDs []string
	if input.UploadResults != nil && input.UploadResults.RequestID != "" {
		// Extract public IDs from successful uploads
		for _, result := range input.UploadResults.Results {
			if result.Error == "" && result.PublicID != "" {
				tempPublicIDs = append(tempPublicIDs, result.PublicID)
			}
		}
		
		// Validate request ID and temporary uploads
		if len(tempPublicIDs) > 0 {
			isValid := utils.Tracker.ValidateAndCleanupRequest(input.UploadResults.RequestID, tempPublicIDs)
			if !isValid {
				utils.HandleError(c, utils.NewValidationError("Invalid or expired upload request"))
				return
			}
		}
	}

	// Create question with initial status
	question := models.Question{
		CourseID:    input.CourseID,
		SessionID:   input.SessionID,
		Type:        input.Type,
		Lecturer:    input.Lecturer,
		TimeAllowed: input.TimeAllowed,
		DocLink:     input.DocLink,
		Tips:        input.Tips,
		Approved:    false, // Default to pending approval
		Downloads:   new(int), // Initialize to 0
		Views:       new(int), // Initialize to 0
		ImageLinks:  []string{}, // Will be populated after image processing
	}

	// Create the question in database first to get the ID
	if err := database.DB.Create(&question).Error; err != nil {
		utils.HandleDatabaseError(c, err)
		return
	}

	// Process images if any were uploaded
	var finalImageURLs []string
	var processingStatus = "processed"
	
	if len(tempPublicIDs) > 0 {
		finalImageURLs, processingStatus = processQuestionImages(tempPublicIDs, question.ID)
	}

	// Update question with final image URLs and processing status
	updateData := map[string]interface{}{
		"image_links": finalImageURLs,
	}

	// If some images failed to process, note it but don't fail the question creation
	if processingStatus != "processed" {
		fmt.Printf("Warning: Some images failed to process for question %s\n", question.ID)
	}

	if err := database.DB.Model(&question).Updates(updateData).Error; err != nil {
		// Log error but don't fail since question was created
		fmt.Printf("Error updating question with image data: %v\n", err)
	}

	// Prepare response
	response := models.QuestionResponse{
		ID:               question.ID,
		CourseID:         question.CourseID,
		SessionID:        question.SessionID,
		Type:             string(question.Type),
		ImageCount:       len(finalImageURLs),
		ProcessingStatus: processingStatus,
		Approved:         question.Approved,
		CreatedAt:        question.CreatedAt.Format(time.RFC3339),
		Message:          "Question created successfully and is pending approval",
	}

	// Add warning message if some images failed
	if processingStatus != "processed" && len(finalImageURLs) > 0 {
		response.Message = "Question created successfully with some image processing issues"
	} else if len(finalImageURLs) == 0 && len(tempPublicIDs) > 0 {
		response.Message = "Question created successfully but all images failed to process"
		response.ProcessingStatus = "failed"
	}

	utils.SuccessResponse(c, response)
}

// processQuestionImages handles concurrent image processing with error resilience
func processQuestionImages(tempPublicIDs []string, questionID string) ([]string, string) {
	const maxConcurrentMoves = 5
	semaphore := make(chan struct{}, maxConcurrentMoves)
	
	results := make([]string, len(tempPublicIDs))
	var wg sync.WaitGroup
	var mu sync.Mutex
	var successCount int

	// Process image moves concurrently
	for i, publicID := range tempPublicIDs {
		wg.Add(1)
		go func(index int, tempPublicID string) {
			defer wg.Done()
			
			// Acquire semaphore
			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			// Move file to permanent location
			finalURL, err := utils.MoveFileToPermanent(tempPublicID, questionID)
			
			mu.Lock()
			if err != nil {
				fmt.Printf("Error moving image %s to permanent location: %v\n", tempPublicID, err)
				results[index] = "" // Mark as failed
			} else {
				results[index] = finalURL
				successCount++
			}
			mu.Unlock()
		}(i, publicID)
	}

	// Wait for all moves to complete
	wg.Wait()

	// Filter out failed moves
	var finalURLs []string
	for _, url := range results {
		if url != "" {
			finalURLs = append(finalURLs, url)
		}
	}

	// Determine processing status
	var status string
	if successCount == len(tempPublicIDs) {
		status = "processed"
	} else if successCount > 0 {
		status = "partial"
	} else {
		status = "failed"
	}

	return finalURLs, status
}

// GetQuestions handles retrieving questions with optional filtering
func GetQuestions(c *gin.Context) {
	var questions []models.Question
	
	query := database.DB
	
	// Add filtering for approved questions only (public endpoint)
	// query = query.Where("approved = ?", true)
	
	// Optional filtering by course
	if courseID := c.Query("courseId"); courseID != "" {
		query = query.Where("course_id = ?", courseID)
	}
	
	// Optional filtering by session
	if sessionID := c.Query("sessionId"); sessionID != "" {
		query = query.Where("session_id = ?", sessionID)
	}
	
	// Optional filtering by type
	if questionType := c.Query("type"); questionType != "" {
		query = query.Where("type = ?", questionType)
	}
	
	// Pagination
	page := getIntQuery(c, "page", 1)
	limit := getIntQuery(c, "limit", 20)
	offset := (page - 1) * limit
	
	if utils.HandleGetResources(c, query.Offset(offset).Limit(limit), &questions) {
		return
	}
}

// GetQuestionByID handles retrieving a single question by ID
func GetQuestionByID(c *gin.Context) {
	id := c.Param("id")
	
	var question models.Question
	if err := database.DB.Preload("Course").Preload("Session").Preload("Uploader").
		Where("id = ? AND approved = ?", id, true).First(&question).Error; err != nil {
		utils.HandleDatabaseErrorWithContext(c, err, "Question")
		return
	}
	
	// Increment view count
	database.DB.Model(&question).Update("views", gorm.Expr("views + 1"))
	
	utils.SuccessResponse(c, question)
}

// Helper function to get integer query parameters
func getIntQuery(c *gin.Context, key string, defaultValue int) int {
	if value := c.Query(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
} 
