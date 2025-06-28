package handlers

import (
	"qb/internal/services"
	"qb/pkg/models"

	"github.com/gin-gonic/gin"
)

// UploadImages handles the image pre-upload endpoint
func UploadImages(c *gin.Context) {
	// Generate request ID for tracking
	requestID := services.GenerateRequestID()

	// Parse multipart form (this is critical for file uploads!)
	err := c.Request.ParseMultipartForm(32 << 20) // 32MB max memory
	if err != nil {
		Res.Invalid(c, err)
		return
	}

	// Get files from form
	form := c.Request.MultipartForm
	if form == nil {
		Res.Invalid(c, "No multipart form data received")
		return
	}

	files := form.File["imageFiles"]

	// Process uploads using service
	response, analysis, err := services.ProcessImageUploads(files, requestID)
	if err != nil {
		Res.Send(c, nil, err)
		return
	}

	// Use helper to handle the response based on analysis
	handleUploadResults(c, analysis, response)
} 

// HandleUploadResults processes upload analysis and sends appropriate error/success response
func handleUploadResults(c *gin.Context, analysis *services.UploadResultAnalysis, response interface{}) {
	if analysis.HasErrors {
		if analysis.SuccessfulUploads == 0 {
			// All uploads failed - determine the primary error type
			if len(analysis.NetworkErrors) > 0 {
				Res.Send(c, nil, models.NewNetworkError(analysis.NetworkErrors))
				return
			} else if len(analysis.UploadErrors) > 0 {
				Res.Send(c, nil, models.NewUploadError(analysis.UploadErrors))
				return
			}
		} else {
			// Partial success - some files uploaded, others failed
			errorDetails := make(map[string]interface{})
			if len(analysis.NetworkErrors) > 0 {
				errorDetails["network_errors"] = analysis.NetworkErrors
			}
			if len(analysis.UploadErrors) > 0 {
				errorDetails["upload_errors"] = analysis.UploadErrors
			}
			errorDetails["successful_uploads"] = analysis.SuccessfulUploads
			errorDetails["total_files"] = analysis.TotalFiles
			
			Res.Send(c, nil, models.NewPartialUploadError(errorDetails))
			return
		}
	}

	Res.Send(c, response, nil)
}
