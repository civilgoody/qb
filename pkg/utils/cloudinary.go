package utils

import (
	"context"
	"fmt"
	"mime/multipart"
	"os"
	"strings"
	"time"

	"github.com/cloudinary/cloudinary-go/v2"
	"github.com/cloudinary/cloudinary-go/v2/api/uploader"
)

var CloudinaryClient *cloudinary.Cloudinary

// InitCloudinary initializes the Cloudinary client using environment variables
func InitCloudinary() error {
	cloudinaryURL := os.Getenv("CLOUDINARY_URL")
	if cloudinaryURL == "" {
		return fmt.Errorf("CLOUDINARY_URL environment variable not set")
	}

	cld, err := cloudinary.NewFromURL(cloudinaryURL)
	if err != nil {
		return fmt.Errorf("failed to initialize Cloudinary: %w", err)
	}

	CloudinaryClient = cld
	return nil
}

// UploadFileToTemp uploads a single image file to temporary folder with request-based tagging
func UploadFileToTemp(fileHeader *multipart.FileHeader, requestID string) (string, string, error) {
	if CloudinaryClient == nil {
		return "", "", fmt.Errorf("cloudinary client not initialized")
	}

	file, err := fileHeader.Open()
	if err != nil {
		return "", "", fmt.Errorf("failed to open file: %w", err)
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
	result, err := CloudinaryClient.Upload.Upload(ctx, file, uploadParams)
	if err != nil {
		return "", "", fmt.Errorf("failed to upload to Cloudinary: %w", err)
	}

	return result.SecureURL, result.PublicID, nil
}

// MoveFileToPermanent moves image from temp folder to permanent folder
func MoveFileToPermanent(tempPublicID, questionID string) (string, error) {
	if CloudinaryClient == nil {
		return "", fmt.Errorf("cloudinary client not initialized")
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
	tempAsset, err := CloudinaryClient.Image(tempPublicID)
	if err != nil {
		return "", fmt.Errorf("failed to get temp image asset: %w", err)
	}
	tempURL, err := tempAsset.String()
	if err != nil {
		return "", fmt.Errorf("failed to generate temp image URL: %w", err)
	}
	
	result, err := CloudinaryClient.Upload.Upload(ctx, tempURL, uploadParams)
	if err != nil {
		return "", fmt.Errorf("failed to move file to permanent location: %w", err)
	}

	// Delete the temporary file
	_, err = CloudinaryClient.Upload.Destroy(ctx, uploader.DestroyParams{
		PublicID: tempPublicID,
	})
	if err != nil {
		// Log error but don't fail the operation since the file was moved successfully
		fmt.Printf("Warning: Failed to delete temp file %s: %v\n", tempPublicID, err)
	}

	return result.SecureURL, nil
}

// extractFilenameFromPublicID extracts the filename part from a Cloudinary public ID
func extractFilenameFromPublicID(publicID string) string {
	parts := strings.Split(publicID, "/")
	return parts[len(parts)-1]
}

// BuildCloudinaryURL constructs a Cloudinary URL from a public ID
func BuildCloudinaryURL(publicID string) string {
	if CloudinaryClient == nil {
		return ""
	}
	
	asset, err := CloudinaryClient.Image(publicID)
	if err != nil {
		return ""
	}
	
	url, err := asset.String()
	if err != nil {
		return ""
	}
	
	return url
} 
