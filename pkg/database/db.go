package database

import (
	"log"
	"qb/pkg/models"
	"qb/pkg/utils"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

func ConnectDB() {
	var err error

	dsn := utils.GetEnvFatal("DATABASE_URL")

	DB, err = gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info), // Log GORM operations
	})

	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	log.Println("Database connected successfully!")

	// Auto-migrate the Todo model
	// This will create the table if it doesn't exist or add missing columns.
	// It will NOT delete columns or change existing column types.
	err = DB.AutoMigrate(models.AllModels...)
	if err != nil {
		log.Fatalf("Failed to auto-migrate database: %v", err)
	}
}
