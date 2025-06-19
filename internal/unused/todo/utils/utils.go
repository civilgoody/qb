package utils

import (
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

// GetEnvFatal retrieves an environment variable or fatally exits if not found or empty.
func GetEnvFatal(key string) string {
	val := os.Getenv(key)
	if val == "" {
		log.Fatalf("%s environment variable is not set or is empty.", key)
	}
	return val
}

// LoadDotEnv loads environment variables from a .env file.
// It logs a warning if the file is not found, but doesn't fatally exit
// as some environments might not use .env files (e.g., production via Docker secrets).
func LoadDotEnv() {
	err := godotenv.Load(".env")
	if err != nil {
		log.Printf("Warning: Could not load .env file (might be ok if using other env sources): %v", err)
	}
}

// HandleInternalServerError sends a 500 Internal Server Error response.
func HandleInternalServerError(c *gin.Context, err error) {
	log.Printf("Internal Server Error: %v", err)
	c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
}
