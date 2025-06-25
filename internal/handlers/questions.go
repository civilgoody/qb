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

// CreateQuestion handles question creation with image finalization
func CreateQuestion(c *gin.Context) {
	// Bind and validate the DTO
	var input models.CreateQuestionDTO
	if utils.BindAndValidate(c, &input) {
		return
	}

	// Validate course and session existence
	if !validateCourseAndSession(c, input.CourseID, input.SessionID) {
		return
	}

	// Process upload results and get temp public IDs
	tempPublicIDs, valid := processUploadResults(c, input.UploadResults)
	if !valid {
		return
	}

	// Generate question ID and process images
	questionID := generateQuestionID(input.CourseID, input.SessionID, input.Type)
	finalImageURLs, processingStatus := processQuestionImages(tempPublicIDs, questionID)

	// Create or update the question
	question, message, created := createOrUpdateQuestion(c, questionID, input, finalImageURLs, processingStatus)
	if question == nil {
		return
	}

	// Build and send response
	sendQuestionResponse(c, question, finalImageURLs, tempPublicIDs, processingStatus, message, created)
}

// validateCourseAndSession validates that the course and session exist
func validateCourseAndSession(c *gin.Context, courseID, sessionID string) bool {
	var course models.Course
	if err := database.DB.Where("id = ?", courseID).First(&course).Error; err != nil {
		utils.HandleDatabaseErrorWithContext(c, err, "Course")
		return false
	}

	var session models.Session
	if err := database.DB.Where("id = ?", sessionID).First(&session).Error; err != nil {
		utils.HandleDatabaseErrorWithContext(c, err, "Session")
		return false
	}

	return true
}

// processUploadResults extracts and validates upload results
func processUploadResults(c *gin.Context, uploadResults *models.UploadResponse) ([]string, bool) {
	var tempPublicIDs []string
	
	if uploadResults != nil && uploadResults.RequestID != "" {
		// Extract public IDs from successful uploads
		for _, result := range uploadResults.Results {
			if result.Error == "" && result.PublicID != "" {
				tempPublicIDs = append(tempPublicIDs, result.PublicID)
			}
		}
		
		// Validate request ID and temporary uploads
		if len(tempPublicIDs) > 0 {
			isValid := utils.Tracker.ValidateAndCleanupRequest(uploadResults.RequestID, tempPublicIDs)
			if !isValid {
				utils.HandleError(c, utils.NewValidationError("Invalid or expired upload request"))
				return nil, false
			}
		}
	}

	return tempPublicIDs, true
}

// createOrUpdateQuestion creates a new question or updates an existing unapproved one
func createOrUpdateQuestion(c *gin.Context, questionID string, input models.CreateQuestionDTO, finalImageURLs []string, processingStatus string) (*models.Question, string, bool) {
	var question models.Question
	
	// Check if question exists and is not approved
	dbResult := database.DB.Where("id = ? AND approved = ?", questionID, false).First(&question)

	if dbResult.Error == nil {
		// Update existing unapproved question
		return updateExistingQuestion(c, &question, finalImageURLs)
	}

	if dbResult.Error != gorm.ErrRecordNotFound {
		utils.HandleDatabaseError(c, dbResult.Error)
		return nil, "", false
	}

	// Create new question
	return createNewQuestion(c, questionID, input, finalImageURLs, processingStatus)
}

// updateExistingQuestion appends images to an existing unapproved question
func updateExistingQuestion(c *gin.Context, question *models.Question, finalImageURLs []string) (*models.Question, string, bool) {
	question.ImageLinks = append(question.ImageLinks, finalImageURLs...)
	if err := database.DB.Save(question).Error; err != nil {
		utils.HandleDatabaseError(c, err)
		return nil, "", false
	}
	return question, "Images appended to existing unapproved question successfully", false
}

// createNewQuestion creates a new question
func createNewQuestion(c *gin.Context, questionID string, input models.CreateQuestionDTO, finalImageURLs []string, processingStatus string) (*models.Question, string, bool) {
	// Get the current user ID from context
	userID, exists := utils.GetCurrentUserID(c)
	if !exists {
		utils.ErrorResponse(c, 401, utils.ErrUnauthorized, "User not found in context")
		return nil, "", false
	}

	question := models.Question{
		ID:               questionID,
		CourseID:         input.CourseID,
		SessionID:        input.SessionID,
		Type:             input.Type,
		Lecturer:         input.Lecturer,
		TimeAllowed:      input.TimeAllowed,
		DocLink:          input.DocLink,
		Tips:             input.Tips,
		Approved:         false,
		Downloads:        new(int),
		Views:            new(int),
		ImageLinks:       finalImageURLs,
		ProcessingStatus: &processingStatus,
		UploaderID:       &userID, // Set the uploader from auth context
	}

	if err := database.DB.Create(&question).Error; err != nil {
		utils.HandleDatabaseError(c, err)
		return nil, "", false
	}

	return &question, "Question created successfully and is pending approval", true
}

// sendQuestionResponse builds and sends the final response
func sendQuestionResponse(c *gin.Context, question *models.Question, finalImageURLs, tempPublicIDs []string, processingStatus, message string, created bool) {
	// Log success
	if len(finalImageURLs) > 0 && len(tempPublicIDs) > 0 {
		action := "created"
		if !created {
			action = "updated"
		}
		fmt.Printf("Successfully %s question %s with %d images, status: %s\n", action, question.ID, len(finalImageURLs), processingStatus)
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

	// Adjust message based on processing status
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
	if len(tempPublicIDs) == 0 {
		return []string{}, "processed"
	}
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
