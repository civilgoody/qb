package services

import (
	"qb/pkg/models"
	"strings"
	"time"

	"github.com/google/uuid"
)

// StartRequestTrackerCleanup starts the cleanup goroutine for expired requests
func StartRequestTrackerCleanup() {
	go startCleanupRoutine()
}

// GenerateRequestID generates a new UUID for tracking upload requests
func GenerateRequestID() string {
	return uuid.New().String()
}

// StoreTemporaryUpload stores a request mapping with 24-hour TTL in database
func StoreTemporaryUpload(requestID string, publicIDs []string) error {
	upload := models.TemporaryUpload{
		RequestID: requestID,
		PublicIDs: strings.Join(publicIDs, ","), // Simple comma-separated string
		ExpiresAt: time.Now().Add(24 * time.Hour),
	}
	
	if err := db.Create(&upload).Error; err != nil {
		return errS.Db(err, "Temporary upload")
	}
	
	return nil
}

// ValidateAndCleanupRequest validates that the request exists and matches the provided public IDs
// If valid, it removes the request from the database and returns true
func ValidateAndCleanupRequest(requestID string, publicIDs []string) bool {
	var upload models.TemporaryUpload
	
	// Find the request
	err := db.Where("request_id = ?", requestID).First(&upload).Error
	if err != nil {
		return false
	}
	
	// Check if request has expired
	if time.Now().After(upload.ExpiresAt) {
		db.Delete(&upload)
		return false
	}
	
	// Convert stored string back to slice and validate
	storedIDs := strings.Split(upload.PublicIDs, ",")
	if !slicesEqual(storedIDs, publicIDs) {
		return false
	}
	
	// Clean up the request after successful validation
	db.Delete(&upload)
	return true
}

// GetRequestInfo retrieves information about a request without removing it
func GetRequestInfo(requestID string) (*models.TemporaryUpload, bool) {
	var upload models.TemporaryUpload
	
	err := db.Where("request_id = ? AND expires_at > ?", requestID, time.Now()).First(&upload).Error
	if err != nil {
		return nil, false
	}
	
	return &upload, true
}

// GetActiveRequestCount returns the number of active requests (for monitoring)
func GetActiveRequestCount() int64 {
	var count int64
	db.Model(&models.TemporaryUpload{}).Where("expires_at > ?", time.Now()).Count(&count)
	return count
}

// startCleanupRoutine runs a periodic cleanup of expired requests
func startCleanupRoutine() {
	ticker := time.NewTicker(1 * time.Hour) // Clean up every hour
	defer ticker.Stop()
	
	for range ticker.C {
		cleanupExpiredRequests()
	}
}

// cleanupExpiredRequests removes expired requests from the database
func cleanupExpiredRequests() {
	db.Where("expires_at < ?", time.Now()).Delete(&models.TemporaryUpload{})
}

// slicesEqual compares two string slices for equality (order matters)
func slicesEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	
	return true
} 
