package services

import (
	"log"
	"qb/pkg/database"
	"qb/pkg/utils"
	"time"

	"github.com/cloudinary/cloudinary-go/v2"
	"github.com/go-playground/validator/v10"
	"gorm.io/gorm"
)

var (
	db   *gorm.DB
	valS *validator.Validate
	cldS *cloudinary.Cloudinary
	errS *ErrorService
	
	// Rate limiter instances - these need to be per-handler since they have different configs
	uploadRateLimiter   *RateLimitService
	generalRateLimiter  *RateLimitService
)

// InitServices initializes all shared service dependencies once
func InitServices() {
	// Initialize database connection
	db = database.DB
	if db == nil {
		log.Fatal("Database connection is nil")
	}

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
	
	// Initialize rate limiters
	InitRateLimiters()

	log.Println("All services initialized successfully")
}

// InitRateLimiters initializes the rate limiting services
func InitRateLimiters() {
	uploadRateLimiter = NewRateLimitService(50, time.Hour)   // 50 uploads per hour
	generalRateLimiter = NewRateLimitService(200, time.Hour) // 200 requests per hour
}

// GetUploadRateLimiter returns the upload rate limiter
func GetUploadRateLimiter() *RateLimitService {
	return uploadRateLimiter
}

// GetGeneralRateLimiter returns the general rate limiter
func GetGeneralRateLimiter() *RateLimitService {
	return generalRateLimiter
}

// GetErrorService returns the shared error service
func GetErrorService() *ErrorService {
	return errS
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
