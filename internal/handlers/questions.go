package handlers

import (
	"qb/internal/services"
	"qb/pkg/models"

	"github.com/gin-gonic/gin"
)

// GetQuestions handles retrieving questions with optional filtering
func GetQuestions(c *gin.Context) {
	// Extract query parameters
	courseID := c.Query("courseId")
	sessionID := c.Query("sessionId")
	questionType := c.Query("type")
	page := services.GetIntQuery(c.Query("page"), 1)
	limit := services.GetIntQuery(c.Query("limit"), 20)

	questions, err := services.GetQuestions(courseID, sessionID, questionType, page, limit)
	Res.Send(c, questions, err)
}

// GetQuestionByID handles retrieving a single question by ID
func GetQuestionByID(c *gin.Context) {
	id := c.Param("id")
	
	question, err := services.GetQuestionByID(id)
	Res.Send(c, question, err)
}

// CreateQuestion handles question creation with image finalization
func CreateQuestion(c *gin.Context) {
	// Bind and validate the DTO
	var input models.CreateQuestionDTO
	if err := c.ShouldBindJSON(&input); err != nil {
		Res.Invalid(c, err)
		return
	}

	// Get the current user ID from context
	userID, err := Auth.GetCurrentUserID(c)
	if err != nil {
		Res.Send(c, nil, err)
		return
	}

	// Create question using service
	question, message, created, err := services.CreateQuestion(input, userID)
	if err != nil {
		Res.Send(c, nil, err)
		return
	}

	// Build response using service
	response := services.BuildQuestionResponse(
		question, 
		question.ImageLinks, 
		extractTempPublicIDs(input.UploadResults), 
		getProcessingStatus(question), 
		message, 
		created,
	)

	Res.Send(c, response, nil)
}

// Helper functions
func extractTempPublicIDs(uploadResults *models.UploadResponse) []string {
	if uploadResults == nil {
		return []string{}
	}
	
	var tempPublicIDs []string
	for _, result := range uploadResults.Results {
		if result.Error == "" && result.PublicID != "" {
			tempPublicIDs = append(tempPublicIDs, result.PublicID)
		}
	}
	return tempPublicIDs
}

func getProcessingStatus(question *models.Question) string {
	if question.ProcessingStatus != nil {
		return *question.ProcessingStatus
	}
	return "processed"
}
