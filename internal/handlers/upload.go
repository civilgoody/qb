package handlers

import (
	"fmt"
	"mime/multipart"
	"qb/pkg/models"
	"qb/pkg/utils"
	"sync"

	"github.com/gin-gonic/gin"
)

// UploadImages handles the image pre-upload endpoint
func UploadImages(c *gin.Context) {
	// Generate request ID for tracking
	requestID := utils.GenerateRequestID()

	// Parse multipart form (this is critical for file uploads!)
	err := c.Request.ParseMultipartForm(32 << 20) // 32MB max memory
	if err != nil {
		utils.HandleValidationError(c, fmt.Errorf("failed to parse multipart form: %w", err))
		return
	}

	// Get files from form
	form := c.Request.MultipartForm
	if form == nil {
		utils.HandleError(c, utils.NewValidationError("No multipart form data received"))
		return
	}

	files := form.File["imageFiles"] // This should match the form field name

	// Validate file count
	if len(files) == 0 {
		utils.HandleError(c, utils.NewValidationError("No files provided"))
		return
	}

	if len(files) > 5 {
		utils.HandleError(c, utils.NewValidationError("Maximum 5 files allowed per request"))
		return
	}

	// Validate each file before processing
	for _, fileHeader := range files {
		if err := utils.ValidateImageFile(fileHeader); err != nil {
			utils.HandleError(c, utils.NewValidationError(fmt.Sprintf("Invalid file '%s': %s", fileHeader.Filename, err.Error())))
			return
		}
	}

	// Use bounded concurrency to prevent overwhelming Cloudinary
	const maxConcurrentUploads = 10
	semaphore := make(chan struct{}, maxConcurrentUploads)
	
	results := make([]models.UploadResult, len(files))
	var wg sync.WaitGroup
	var mu sync.Mutex

	// Process uploads concurrently
	for i, fileHeader := range files {
		wg.Add(1)
		go func(index int, file *multipart.FileHeader) {
			defer wg.Done()
			
			// Acquire semaphore
			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			result := models.UploadResult{
				OriginalFilename: file.Filename,
			}

			// Upload file to Cloudinary
			_, publicID, err := utils.UploadFileToTemp(file, requestID)
			if err != nil {
				result.Error = err.Error()
			} else {
				result.PublicID = publicID
			}

			// Store result thread-safely
			mu.Lock()
			results[index] = result
			mu.Unlock()
		}(i, fileHeader)
	}

	// Wait for all uploads to complete
	wg.Wait()

	// Collect successful uploads for tracking
	var successfulPublicIDs []string
	for _, result := range results {
		if result.Error == "" {
			successfulPublicIDs = append(successfulPublicIDs, result.PublicID)
		}
	}

	// Store successful uploads in request tracker
	if len(successfulPublicIDs) > 0 {
		if err := utils.Tracker.StoreTemporaryUpload(requestID, successfulPublicIDs); err != nil {
			// Log the error but don't fail the upload response since files were uploaded successfully
			fmt.Printf("Warning: Failed to store request tracking: %v\n", err)
		}
	}

	// Analyze upload results using the helper
	analysis := utils.AnalyzeUploadResults(results)

	// Prepare response
	response := models.UploadResponse{
		RequestID: requestID,
		Results:   results,
		Success:   !analysis.HasErrors && len(successfulPublicIDs) > 0,
	}

	// Use helper to handle the response based on analysis
	utils.HandleUploadResults(c, analysis, response)
} 
