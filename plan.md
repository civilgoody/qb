# Feature Implementation Plan: Staged Image Uploads for Public Past Question Library

## Overview

This plan outlines the implementation of a staged image upload feature for a public past question library. Users will first upload images to a temporary Cloudinary folder via a dedicated backend endpoint. Upon successful temporary uploads, the frontend will receive temporary image URLs and a request ID. When the user submits the main question form, these temporary URLs and request ID will be sent to the backend, which will validate the request, concurrently move/finalize the images to a permanent Cloudinary location, and save the question record with approval status pending.

## Prerequisites

1. **Cloudinary Account & Credentials:**

   - Ensure you have a Cloudinary account.
   - Retrieve your `CLOUDINARY_CLOUD_NAME`, `CLOUDINARY_API_KEY`, and `CLOUDINARY_API_SECRET` from your Cloudinary dashboard.

2. **Install Cloudinary Go SDK:**

   - Run the following command in your terminal:
     ```bash
     go get github.com/cloudinary/cloudinary-go/v2
     ```

3. **Configure Environment Variables:**

   - Add your Cloudinary credentials to your `.env` file at the project root:
     ```ini
     CLOUDINARY_CLOUD_NAME=your_cloud_name
     CLOUDINARY_API_KEY=your_api_key
     CLOUDINARY_API_SECRET=your_api_secret
     ```

4. **Rate Limiting Setup:**
   - Install rate limiting middleware:
     ```bash
     go get github.com/gin-contrib/limiter
     ```

## Phase 1: Image Pre-Upload Endpoint (`/api/v1/upload-images`)

**Purpose:** To allow users to upload images and receive immediate feedback, storing images temporarily on Cloudinary with request tracking for validation.

### Backend Components:

1. **Request Tracking (`pkg/utils/request_tracker.go`)**

   - **In-Memory Store with TTL:** Use `sync.Map` or Redis to store temporary upload requests.
   - **`GenerateRequestID() string`:** Generate UUID for each upload request.
   - **`StoreTemporaryUpload(requestID string, publicIDs []string)`:** Store request mapping with 24-hour TTL.
   - **`ValidateAndCleanupRequest(requestID string, publicIDs []string) bool`:** Validate request exists and matches, then remove from store.

2. **Cloudinary Utility (`pkg/utils/cloudinary.go`)**

   - **Initialization:** Add `InitCloudinary()` function to initialize the Cloudinary client.
   - **`UploadFileToTemp(fileHeader *multipart.FileHeader, requestID string) (string, string, error)`:**
     - **Purpose:** Uploads a single image file to temporary folder with request-based tagging.
     - **Tagging:** Use tags like `temp_upload`, `req_${requestID}`, `expires_${timestamp}` for auto-cleanup.
     - **Input:** `*multipart.FileHeader` and `requestID`.
     - **Output:** The temporary `secure_url`, `public_id`, and error.

3. **Input/Output DTOs (`pkg/models/upload-dto.go`)**

   ```go
   package models

   import "mime/multipart"

   type UploadImagesDTO struct {
       ImageFiles []*multipart.FileHeader `form:"imageFiles" binding:"required,max=5"`
   }

   type UploadResult struct {
       OriginalFilename string `json:"originalFilename"`
       TempURL          string `json:"tempUrl,omitempty"`
       PublicID         string `json:"publicId,omitempty"`
       Error            string `json:"error,omitempty"`
   }

   type UploadResponse struct {
       RequestID string         `json:"requestId"`
       Results   []UploadResult `json:"results"`
       Success   bool           `json:"success"`
   }
   ```

4. **Rate Limiting Middleware (`pkg/middleware/rate_limit.go`)**

   - **IP-based Rate Limiting:** 50 uploads per hour per IP.
   - **File Validation:** Max 5 files, 10MB each, allowed MIME types (image/jpeg, image/png, image/webp).

5. **Handler (`internal/handlers/upload.go`)**

   - **Implement `UploadImages(c *gin.Context)`:**
     - Generate `requestID` using `utils.GenerateRequestID()`.
     - Bind `UploadImagesDTO` and validate files.
     - **Bounded Concurrency:** Use worker pool pattern (max 10 concurrent uploads).
     - Launch goroutines for each file upload with `cloudinary.UploadFileToTemp()`.
     - Collect results and store successful uploads in request tracker.
     - Return `UploadResponse` with `requestID` and individual results.

6. **Routing (`internal/routes/r.go`)**
   ```go
   // Add rate limiting middleware
   v1.POST("/upload-images", middleware.RateLimit(), handlers.UploadImages)
   ```

## Phase 2: Question Submission & Image Finalization (`/api/v1/questions` POST)

**Purpose:** To receive the main question form data with request validation, move images to permanent location, and create the Question record with pending approval status.

### Backend Components:

1. **Cloudinary Utility Enhancement (`pkg/utils/cloudinary.go`)**

   - **`MoveFileToPermanent(tempPublicID, questionID string) (string, error)`:**
     - **Purpose:** Moves image from temp folder to permanent folder: `qb_questions/${questionID}/`.
     - **Tagging:** Replace temp tags with `permanent`, `question_${questionID}`.
     - **Transformations:** Apply auto-format, auto-quality, and generate eager thumbnails.
     - **Input:** Temporary `public_id` and `questionID`.
     - **Output:** Final permanent `secure_url` or error.

2. **Updated Question Input DTO (`pkg/models/question-dto.go`)**

   ```go
   type CreateQuestionDTO struct {
       // ... existing fields ...
       RequestID    string   `json:"requestId" binding:"required,uuid"`
       TempImageURLs []string `json:"tempImageUrls,omitempty" validate:"omitempty,dive,url"`
       TempPublicIDs []string `json:"tempPublicIds,omitempty"`
       // ... other fields ...
   }
   ```

3. **Enhanced Question Schema Considerations**

   ```go
   type Question struct {
       // ... existing fields from your schema ...

       // Consider adding these for better tracking
       ImageCount       int      `gorm:"default:0" json:"imageCount"`
       ProcessingStatus string   `gorm:"default:'pending'" json:"processingStatus"` // pending, processed, failed
       SubmittedAt      time.Time `gorm:"autoCreateTime" json:"submittedAt"`

       // Your existing Approved field covers moderation
       Approved bool `gorm:"default:false" json:"approved"`
   }
   ```

4. **Handler (`internal/handlers/qb.go`)**
   - **Implement `CreateQuestion(c *gin.Context)`:**
     - Bind `CreateQuestionDTO` and validate.
     - **Request Validation:** Use `request_tracker.ValidateAndCleanupRequest()` to verify requestID and publicIDs match.
     - **Bounded Concurrency:** Use worker pool for concurrent image moves.
     - Loop through `input.TempPublicIDs`, launching goroutines for `cloudinary.MoveFileToPermanent()`.
     - **Error Handling Policy:**
       - If all moves fail: Save question without images, log errors.
       - If partial success: Save question with successful images, log failures.
       - If all succeed: Save question with all images.
     - Create `models.Question` with `Approved: false` (pending moderation).
     - Set `ImageCount` and `ProcessingStatus` based on results.
     - Persist to database and return success response.

## Phase 3: Cleanup & Maintenance

### 1. **Automatic Cloudinary Cleanup**

- **Cloudinary Auto-Delete:** Configure Cloudinary to auto-delete resources tagged with `temp_upload` after 48 hours.
- **Backup Cleanup Job:** Optional daily cron job to clean up any missed temporary uploads.

### 2. **Request Tracker Cleanup**

- **TTL-based Cleanup:** If using Redis, set TTL to 24 hours.
- **In-Memory Cleanup:** Run periodic cleanup goroutine to remove expired entries.

### 3. **Monitoring & Health Checks**

- **Cloudinary Health Check:** Periodic connectivity test to Cloudinary API.
- **Metrics Collection:** Track upload success rates, processing times, storage usage.
- **Error Logging:** Structured logging for all Cloudinary operations and failures.

## Security Considerations for Public App

1. **Rate Limiting:**

   - IP-based limits: 50 uploads/hour, 200 requests/hour per IP.
   - Global limits: Monitor for sudden spikes in usage.

2. **Content Validation:**

   - File type validation using actual MIME type detection.
   - File size limits (10MB per file, 50MB total per request).
   - Basic image validation (valid image headers).

3. **Spam Prevention:**

   - Optional: Simple CAPTCHA or proof-of-work challenge.
   - Request ID expiration to prevent replay attacks.
   - IP-based temporary bans for abuse.

4. **Moderation Queue:**
   - All questions default to `Approved: false`.
   - Admin interface for reviewing and approving questions.
   - Automated content scanning for inappropriate material.

## Testing Strategy

1. **Unit Tests:**

   - Test `cloudinary.UploadFileToTemp()` and `cloudinary.MoveFileToPermanent()`.
   - Test request tracking utility functions.
   - Test DTO validation and binding.

2. **Integration Tests:**

   - Test complete upload flow with actual files.
   - Test question creation with various image scenarios.
   - Test rate limiting and error handling.

3. **Load Tests:**

   - Test concurrent uploads from multiple IPs.
   - Test Cloudinary API limits and worker pool efficiency.
   - Test cleanup mechanisms under load.

4. **End-to-End Tests:**
   - Test complete user flow: upload → create question → moderation.
   - Test error scenarios: network failures, invalid files, expired requests.

## Performance Optimizations

1. **Worker Pools:** Limit concurrent Cloudinary operations to prevent rate limit hits.
2. **CDN Optimization:** Use Cloudinary's CDN features for faster image delivery.
3. **Eager Transformations:** Pre-generate thumbnails and common sizes.
4. **Caching:** Cache frequently accessed questions and images.
5. **Database Indexing:** Index on `CourseID`, `SessionID`, `Approved`, `CreatedAt` for fast queries.

This plan provides a robust, scalable solution for a public past question library while maintaining simplicity and avoiding unnecessary complexity around user management.
