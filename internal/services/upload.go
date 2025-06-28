package services

import (
	"fmt"
	"mime/multipart"
	"qb/pkg/models"
	"qb/pkg/utils"
	"sync"
)

// ProcessImageUploads handles the core logic for uploading multiple images
func ProcessImageUploads(files []*multipart.FileHeader, requestID string) (models.UploadResponse, *UploadResultAnalysis, error) {
	// Validate file count
	if len(files) == 0 {
		return models.UploadResponse{}, nil, errS.Invalid("No files provided")
	}

	if len(files) > 5 {
		return models.UploadResponse{}, nil, errS.Invalid("Maximum 5 files allowed per request")
	}

	// Validate each file before processing
	for _, fileHeader := range files {
		if err := ValidateImageFile(fileHeader); err != nil {
			return models.UploadResponse{}, nil, errS.Invalid(fmt.Sprintf("Invalid file '%s': %s", fileHeader.Filename, err.Error()))
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
			_, publicID, err := UploadFileToTemp(file, requestID)
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
		if err := StoreTemporaryUpload(requestID, successfulPublicIDs); err != nil {
			// Log the error but don't fail the upload response since files were uploaded successfully
			fmt.Printf("Warning: Failed to store request tracking: %v\n", err)
		}
	}

	// Analyze upload results - do this only once here
	analysis := AnalyzeUploadResults(results)

	// Prepare response
	response := models.UploadResponse{
		RequestID: requestID,
		Results:   results,
		Success:   !analysis.HasErrors && len(successfulPublicIDs) > 0,
	}

	return response, analysis, nil
}

type UploadResultAnalysis struct {
	HasErrors           bool
	NetworkErrors       []string
	UploadErrors        []string
	SuccessfulUploads   int
	TotalFiles          int
}

// AnalyzeUploadResults categorizes upload errors and determines appropriate response
func AnalyzeUploadResults(results []models.UploadResult) *UploadResultAnalysis {
	analysis := &UploadResultAnalysis{
		NetworkErrors: make([]string, 0),
		UploadErrors:  make([]string, 0),
		TotalFiles:    len(results),
	}

	for _, result := range results {
		errorMsg := result.Error
		if errorMsg != "" {
			analysis.HasErrors = true
			filename := result.OriginalFilename
			
			// Categorize the error type
			if utils.IsNetworkError(result.Error) {
				analysis.NetworkErrors = append(analysis.NetworkErrors, fmt.Sprintf("%s: %s", filename, errorMsg))
			} else {
				analysis.UploadErrors = append(analysis.UploadErrors, fmt.Sprintf("%s: %s", filename, errorMsg))
			}
		} else {
			analysis.SuccessfulUploads++
		}
	}

	return analysis
}
