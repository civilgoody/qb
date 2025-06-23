package utils

import (
	"fmt"
	"qb/pkg/models"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// TemporaryUpload represents a temporary upload request
type TemporaryUpload struct {
	PublicIDs []string
	ExpiresAt time.Time
}

// RequestTracker manages temporary upload requests using database persistence
type RequestTracker struct {
	db *gorm.DB
}

var Tracker *RequestTracker

// InitRequestTracker initializes the global request tracker
func InitRequestTracker(db *gorm.DB) {
	Tracker = &RequestTracker{db: db}
	
	// Start cleanup goroutine
	go Tracker.startCleanupRoutine()
}

// GenerateRequestID generates a new UUID for tracking upload requests
func GenerateRequestID() string {
	return uuid.New().String()
}

// StoreTemporaryUpload stores a request mapping with 24-hour TTL in database
func (rt *RequestTracker) StoreTemporaryUpload(requestID string, publicIDs []string) error {
	upload := models.TemporaryUpload{
		RequestID: requestID,
		PublicIDs: strings.Join(publicIDs, ","), // Simple comma-separated string
		ExpiresAt: time.Now().Add(24 * time.Hour),
	}
	
	if err := rt.db.Create(&upload).Error; err != nil {
		return fmt.Errorf("failed to store temporary upload: %w", err)
	}
	
	return nil
}

// ValidateAndCleanupRequest validates that the request exists and matches the provided public IDs
// If valid, it removes the request from the database and returns true
func (rt *RequestTracker) ValidateAndCleanupRequest(requestID string, publicIDs []string) bool {
	var upload models.TemporaryUpload
	
	// Find the request
	err := rt.db.Where("request_id = ?", requestID).First(&upload).Error
	if err != nil {
		return false
	}
	
	// Check if request has expired
	if time.Now().After(upload.ExpiresAt) {
		rt.db.Delete(&upload)
		return false
	}
	
	// Convert stored string back to slice and validate
	storedIDs := strings.Split(upload.PublicIDs, ",")
	if !slicesEqual(storedIDs, publicIDs) {
		return false
	}
	
	// Clean up the request after successful validation
	rt.db.Delete(&upload)
	return true
}

// GetRequestInfo retrieves information about a request without removing it
func (rt *RequestTracker) GetRequestInfo(requestID string) (*models.TemporaryUpload, bool) {
	var upload models.TemporaryUpload
	
	err := rt.db.Where("request_id = ? AND expires_at > ?", requestID, time.Now()).First(&upload).Error
	if err != nil {
		return nil, false
	}
	
	return &upload, true
}

// startCleanupRoutine runs a periodic cleanup of expired requests
func (rt *RequestTracker) startCleanupRoutine() {
	ticker := time.NewTicker(1 * time.Hour) // Clean up every hour
	defer ticker.Stop()
	
	for range ticker.C {
		rt.cleanupExpiredRequests()
	}
}

// cleanupExpiredRequests removes expired requests from the database
func (rt *RequestTracker) cleanupExpiredRequests() {
	rt.db.Where("expires_at < ?", time.Now()).Delete(&models.TemporaryUpload{})
}

// GetActiveRequestCount returns the number of active requests (for monitoring)
func (rt *RequestTracker) GetActiveRequestCount() int64 {
	var count int64
	rt.db.Model(&models.TemporaryUpload{}).Where("expires_at > ?", time.Now()).Count(&count)
	return count
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
