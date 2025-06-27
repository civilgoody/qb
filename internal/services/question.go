package services

import (
	"fmt"
	"qb/pkg/models"
	"qb/pkg/utils"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/go-playground/validator/v10"
	"gorm.io/gorm"
)

// QuestionService handles question-specific business logic
type QuestionService struct {
	db       *gorm.DB
	validate *validator.Validate
	err      *ErrorService
}

// NewQuestionService creates a new question service instance
func NewQuestionService(db *gorm.DB, validate *validator.Validate) *QuestionService {
	return &QuestionService{
		db:       db,
		validate: validate,
		err:      NewErrorService(),
	}
}

// GetQuestions retrieves questions with optional filtering
func (s *QuestionService) GetQuestions(courseID, sessionID, questionType string, page, limit int) ([]models.Question, error) {
	var questions []models.Question
	
	query := s.db
	
	// Add filtering for approved questions only (public endpoint)
	// query = query.Where("approved = ?", true)
	
	// Optional filtering by course
	if courseID != "" {
		query = query.Where("course_id = ?", courseID)
	}
	
	// Optional filtering by session
	if sessionID != "" {
		query = query.Where("session_id = ?", sessionID)
	}
	
	// Optional filtering by type
	if questionType != "" {
		query = query.Where("type = ?", questionType)
	}
	
	// Pagination
	offset := (page - 1) * limit
	
	if err := query.Offset(offset).Limit(limit).Find(&questions).Error; err != nil {
		return nil, s.err.Db(err)
	}

	return questions, nil
}

// GetQuestionByID retrieves a single question by ID and increments view count
func (s *QuestionService) GetQuestionByID(id string) (*models.Question, error) {
	var question models.Question
	if err := s.db.Preload("Course").Preload("Session").Preload("Uploader").
		Where("id = ? AND approved = ?", id, true).First(&question).Error; err != nil {
		return nil, s.err.Db(err, "Question")
	}
	
	// Increment view count
	s.db.Model(&question).Update("views", gorm.Expr("views + 1"))
	
	return &question, nil
}

// CreateQuestion handles question creation with image finalization
func (s *QuestionService) CreateQuestion(input models.CreateQuestionDTO, userID string) (*models.Question, string, bool, error) {
	// Validate the DTO
	if err := s.validate.Struct(input); err != nil {
		return nil, "", false, s.err.Invalid(err)
	}

	// Validate course and session existence
	if err := s.validateCourseAndSession(input.CourseID, input.SessionID); err != nil {
		return nil, "", false, err
	}

	// Process upload results and get temp public IDs
	tempPublicIDs, err := s.processUploadResults(input.UploadResults)
	if err != nil {
		return nil, "", false, err
	}

	// Generate question ID and process images
	questionID := s.generateQuestionID(input.CourseID, input.SessionID, input.Type)
	finalImageURLs, processingStatus := s.processQuestionImages(tempPublicIDs, questionID)

	// Create or update the question
	question, message, created, err := s.createOrUpdateQuestion(questionID, input, finalImageURLs, processingStatus, userID)
	if err != nil {
		return nil, "", false, err
	}

	return question, message, created, nil
}

// validateCourseAndSession validates that the course and session exist
func (s *QuestionService) validateCourseAndSession(courseID, sessionID string) error {
	var course models.Course
	if err := s.db.Where("id = ?", courseID).First(&course).Error; err != nil {
		return s.err.Db(err, "Course")
	}

	var session models.Session
	if err := s.db.Where("id = ?", sessionID).First(&session).Error; err != nil {
		return s.err.Db(err, "Session")
	}

	return nil
}

// processUploadResults extracts and validates upload results
func (s *QuestionService) processUploadResults(uploadResults *models.UploadResponse) ([]string, error) {
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
				return nil, s.err.Invalid("Invalid or expired upload request")
			}
		}
	}

	return tempPublicIDs, nil
}

// createOrUpdateQuestion creates a new question or updates an existing unapproved one
func (s *QuestionService) createOrUpdateQuestion(questionID string, input models.CreateQuestionDTO, finalImageURLs []string, processingStatus, userID string) (*models.Question, string, bool, error) {
	var question models.Question
	
	// Check if question exists and is not approved
	dbResult := s.db.Where("id = ? AND approved = ?", questionID, false).First(&question)

	if dbResult.Error == nil {
		// Update existing unapproved question
		return s.updateExistingQuestion(&question, finalImageURLs)
	}

	if dbResult.Error != gorm.ErrRecordNotFound {
		return nil, "", false, s.err.Db(dbResult.Error)
	}

	// Create new question
	return s.createNewQuestion(questionID, input, finalImageURLs, processingStatus, userID)
}

// updateExistingQuestion appends images to an existing unapproved question
func (s *QuestionService) updateExistingQuestion(question *models.Question, finalImageURLs []string) (*models.Question, string, bool, error) {
	question.ImageLinks = append(question.ImageLinks, finalImageURLs...)
	if err := s.db.Save(question).Error; err != nil {
		return nil, "", false, s.err.Db(err)
	}
	return question, "Images appended to existing unapproved question successfully", false, nil
}

// createNewQuestion creates a new question
func (s *QuestionService) createNewQuestion(questionID string, input models.CreateQuestionDTO, finalImageURLs []string, processingStatus, userID string) (*models.Question, string, bool, error) {
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
		UploaderID:       &userID,
	}

	if err := s.db.Create(&question).Error; err != nil {
		return nil, "", false, s.err.Db(err)
	}

	return &question, "Question created successfully and is pending approval", true, nil
}

// processQuestionImages handles concurrent image processing with error resilience
func (s *QuestionService) processQuestionImages(tempPublicIDs []string, questionID string) ([]string, string) {
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

// generateQuestionID generates an ID for the question based on course, session, and type
func (s *QuestionService) generateQuestionID(courseID string, sessionID string, questionType models.QuestionType) string {
	// Convert courseID to lowercase for the ID
	lowerCourseID := strings.ToLower(courseID)

	// Get the first letter of the question type, converted to lowercase
	typeInitial := strings.ToLower(string(questionType[0]))

	// Format: courseId-sessionId-typeInitial
	return fmt.Sprintf("%s-%s-%s", lowerCourseID, sessionID, typeInitial)
}

// BuildQuestionResponse builds the response for question creation
func (s *QuestionService) BuildQuestionResponse(question *models.Question, finalImageURLs, tempPublicIDs []string, processingStatus, message string, created bool) models.QuestionResponse {
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

	return response
}

// Helper function to get integer query parameters with default value
func GetIntQuery(value string, defaultValue int) int {
	if value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
} 
