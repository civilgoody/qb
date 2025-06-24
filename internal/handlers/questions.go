package handlers

import (
	"fmt"
	"qb/pkg/database"
	"qb/pkg/models"
	"qb/pkg/utils"
	"strconv"
	"strings"
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

	// Generate question ID first so we can use it for image processing
	questionID := generateQuestionID(input.CourseID, input.SessionID, input.Type)
	
	// Process images first if any were uploaded
	var finalImageURLs []string
	var processingStatus = "processed"
	
	if len(tempPublicIDs) > 0 {
		finalImageURLs, processingStatus = processQuestionImages(tempPublicIDs, questionID)
	}

	var question models.Question
	message := "Question created successfully and is pending approval"

	// Check if a question with this ID already exists and is not approved
	dbResult := database.DB.Where("id = ? AND approved = ?", questionID, false).First(&question)
	fmt.Println(dbResult.Error)
	

	if dbResult.Error == nil {
		// Question exists and is not approved, append image links
		question.ImageLinks = append(question.ImageLinks, finalImageURLs...)
		if err := database.DB.Save(&question).Error; err != nil {
			utils.HandleDatabaseError(c, err)
			return
		}
		message = "Images appended to existing unapproved question successfully"

		// If we used a temp ID for images, update them to use the real question ID
		if len(finalImageURLs) > 0 && len(tempPublicIDs) > 0 {
			fmt.Printf("Successfully updated question %s with %d images, status: %s\n", question.ID, len(finalImageURLs), processingStatus)
		}

		// Prepare response
		response := models.QuestionResponse{
			ID:               question.ID,
			CourseID:         question.CourseID,
			SessionID:        question.SessionID,
			Type:             question.Type,
			ImageCount:       len(question.ImageLinks),
			ImageLinks:       question.ImageLinks,
			ProcessingStatus: processingStatus,
			Approved:         question.Approved,
			CreatedAt:        question.CreatedAt.Format(time.RFC3339),
			Message:          message,
		}

		// Add warning message if some images failed
		if processingStatus != "processed" && len(finalImageURLs) > 0 {
			response.Message = "Question created successfully with some image processing issues"
		} else if len(finalImageURLs) == 0 && len(tempPublicIDs) > 0 {
			response.Message = "Question created successfully but all images failed to process"
			response.ProcessingStatus = "failed"
		}

		utils.SuccessResponse(c, response)
		return
	}

	if dbResult.Error != gorm.ErrRecordNotFound {
		utils.HandleDatabaseError(c, dbResult.Error)
		return
	}

	// Question does not exist or is already approved, create a new one
	question = models.Question{
		ID:               questionID,
		CourseID:         input.CourseID,
		SessionID:        input.SessionID,
		Type:             input.Type,
		Lecturer:         input.Lecturer,
		TimeAllowed:      input.TimeAllowed,
		DocLink:          input.DocLink,
		Tips:             input.Tips,
		Approved:         false, // Default to pending approval
		Downloads:        new(int), // Initialize to 0
		Views:            new(int), // Initialize to 0
		ImageLinks:       finalImageURLs,
		ProcessingStatus: &processingStatus,
	}

	// Create the question in database with all data
	if err := database.DB.Create(&question).Error; err != nil {
		utils.HandleDatabaseError(c, err)
		return
	}

	// If we used a temp ID for images, update them to use the real question ID
	if len(finalImageURLs) > 0 && len(tempPublicIDs) > 0 {
		fmt.Printf("Successfully created question %s with %d images, status: %s\n", question.ID, len(finalImageURLs), processingStatus)
	}

	// Prepare response
	response := models.QuestionResponse{
		ID:               question.ID,
		CourseID:         question.CourseID,
		SessionID:        question.SessionID,
		Type:             question.Type,
		ImageCount:       len(question.ImageLinks), // Use question.ImageLinks which now includes all images
		ImageLinks:       question.ImageLinks,
		ProcessingStatus: processingStatus,
		Approved:         question.Approved,
		CreatedAt:        question.CreatedAt.Format(time.RFC3339),
		Message:          message,
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
	status := "failed"
	
	if successCount == len(tempPublicIDs) {
		status = "processed"
	} else if successCount > 0 {
		status = "partial"
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

// generateQuestionID generates an ID for the question based on course, session, and type
func generateQuestionID(courseID string, sessionID string, questionType models.QuestionType) string {
	// Convert courseID to lowercase for the ID
	lowerCourseID := strings.ToLower(courseID)

	// Get the first letter of the question type, converted to lowercase
	typeInitial := strings.ToLower(string(questionType[0]))

	// Format: courseId-sessionId-typeInitial
	return fmt.Sprintf("%s-%s-%s", lowerCourseID, sessionID, typeInitial)
}
