package services

import (
	"context"
	"fmt"
	"mime/multipart"
	"net/http"
	"qb/pkg/models"
	"strings"
	"time"

	"github.com/cloudinary/cloudinary-go/v2/api/uploader"
)

// UploadFileToTemp uploads a single image file to temporary folder with request-based tagging
func UploadFileToTemp(fileHeader *multipart.FileHeader, requestID string) (string, string, error) {
	if cldS == nil {
		return "", "", models.ErrInternal
	}

	file, err := fileHeader.Open()
	if err != nil {
		return "", "", errS.Invalid(fmt.Sprintf("Failed to open file: %v", err))
	}
	defer file.Close()

	// Generate expiration timestamp (24 hours from now)
	expirationTime := time.Now().Add(24 * time.Hour).Unix()

	// Upload parameters with tagging for auto-cleanup
	uploadParams := uploader.UploadParams{
		Folder: "qb_temp_uploads",
		Tags: []string{
			"temp_upload",
			fmt.Sprintf("req_%s", requestID),
			fmt.Sprintf("expires_%d", expirationTime),
		},
		ResourceType: "image",
		Transformation: "f_auto,q_auto", // Auto-detect format and optimize quality
	}

	ctx := context.Background()
	result, err := cldS.Upload.Upload(ctx, file, uploadParams)
	if err != nil {
		return "", "", fmt.Errorf("failed to upload to Cloudinary: %w", err)
	}

	return result.SecureURL, result.PublicID, nil
}

// MoveFileToPermanent moves image from temp folder to permanent folder
func MoveFileToPermanent(tempPublicID, questionID string) (string, error) {
	if cldS == nil {
		return "", models.ErrInternal
	}

	ctx := context.Background()

	// Generate new public ID for permanent location
	newPublicID := fmt.Sprintf("qb_questions/%s/%s", questionID, extractFilenameFromPublicID(tempPublicID))

	// Copy the file to permanent location with new upload
	uploadParams := uploader.UploadParams{
		PublicID: newPublicID,
		Tags: []string{
			"permanent",
			fmt.Sprintf("question_%s", questionID),
		},
		Transformation: "f_auto,q_auto",
		ResourceType: "image",
	}

	// Get the temporary file URL
	tempAsset, err := cldS.Image(tempPublicID)
	if err != nil {
		return "", fmt.Errorf("failed to get temp image asset: %w", err)
	}
	tempURL, err := tempAsset.String()
	if err != nil {
		return "", fmt.Errorf("failed to generate temp image URL: %w", err)
	}
	
	result, err := cldS.Upload.Upload(ctx, tempURL, uploadParams)
	if err != nil {
		return "", fmt.Errorf("failed to move file to permanent location: %w", err)
	}

	// Delete the temporary file
	_, err = cldS.Upload.Destroy(ctx, uploader.DestroyParams{
		PublicID: tempPublicID,
	})
	if err != nil {
		// Log error but don't fail the operation since the file was moved successfully
		fmt.Printf("Warning: Failed to delete temp file %s: %v\n", tempPublicID, err)
	}

	return result.SecureURL, nil
}

// BuildCloudinaryURL constructs a Cloudinary URL from a public ID
func BuildCloudinaryURL(publicID string) (string, error) {
	if cldS == nil {
		return "", models.ErrInternal
	}
	
	asset, err := cldS.Image(publicID)
	if err != nil {
		return "", errS.Invalid(fmt.Sprintf("Failed to create image asset: %v", err))
	}
	
	url, err := asset.String()
	if err != nil {
		return "", errS.Invalid(fmt.Sprintf("Failed to generate image URL: %v", err))
	}
	
	return url, nil
}

// ValidateImageFile validates uploaded image files
func ValidateImageFile(fileHeader *multipart.FileHeader) error {
	// Check file size (10MB limit)
	if fileHeader.Size > 10*1024*1024 {
		return errS.Invalid("File size exceeds 10MB limit")
	}
	
	// Check MIME type
	file, err := fileHeader.Open()
	if err != nil {
		return errS.Invalid(fmt.Sprintf("Failed to open file: %v", err))
	}
	defer file.Close()
	
	// Read first 512 bytes to detect content type
	buffer := make([]byte, 512)
	_, err = file.Read(buffer)
	if err != nil {
		return errS.Invalid(fmt.Sprintf("Failed to read file content: %v", err))
	}
	
	contentType := http.DetectContentType(buffer)
	allowedTypes := map[string]bool{
		"image/jpeg": true,
		"image/jpg":  true,
		"image/png":  true,
		"image/webp": true,
	}
	
	if !allowedTypes[contentType] {
		return errS.Invalid(fmt.Sprintf("Unsupported file type: %s. Allowed types: JPEG, PNG, WebP", contentType))
	}
	
	return nil
}

// extractFilenameFromPublicID extracts the filename part from a Cloudinary public ID
func extractFilenameFromPublicID(publicID string) string {
	parts := strings.Split(publicID, "/")
	return parts[len(parts)-1]
}

// DetectContentType mimics http.DetectContentType but can be overridden for testing
var DetectContentType = func(data []byte) string {
	// This would normally be http.DetectContentType(data)
	// But we'll keep it simple for now and infer from file extension
	return "image/jpeg" // Default for now, can be enhanced
} 
