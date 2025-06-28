package services

import (
	"log"
	"qb/pkg/database"
	"qb/pkg/utils"

	"github.com/cloudinary/cloudinary-go/v2"
	"github.com/go-playground/validator/v10"
	"gorm.io/gorm"
)

var (
	db   *gorm.DB
	valS *validator.Validate
	cldS *cloudinary.Cloudinary
	errS *ErrorService
)

// InitServices initializes all shared dependencies once
func InitServices() {
	// Initialize validator
	valS = validator.New()

	// Initialize cloudinary
	cloudinaryURL := utils.GetEnvFatal("CLOUDINARY_URL")
	cld, err := cloudinary.NewFromURL(cloudinaryURL)
	if err != nil {
		log.Fatalf("Failed to initialize Cloudinary: %v", err)
	}
	cldS = cld

	// Initialize database (use existing connection)
	database.ConnectDB()
	db = database.DB

	// Initialize error service
	errS = NewErrorService()
	
	// Start background cleanup routines
	StartRequestTrackerCleanup()
}

// Getter functions for external packages if needed
func GetDB() *gorm.DB {
	return db
}

func GetValidator() *validator.Validate {
	return valS
}

func GetCloudinary() *cloudinary.Cloudinary {
	return cldS
}

func GetErrorService() *ErrorService {
	return errS
}
