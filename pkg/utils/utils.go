package utils

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

func GetEnvFatal(key string) string {
	value := os.Getenv(key)
	if value == "" {
		log.Fatalf("Environment variable %s is not set", key)
	}

	return value
}

func LoadDotEnv() {
	err := godotenv.Load(".env")
	if err != nil {
		log.Printf("Warning: Could not load .env file (might be ok if using other env sources): %v", err)
	}
}
