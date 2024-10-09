package database

import (
	"log"
	"os"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// DB is a global variable for database connection
var DB *gorm.DB

// Connect establishes a connection to the database using credentials from environment variables
func Connect() {
	dsn := os.Getenv("DB_USER") + ":" +
		os.Getenv("DB_PASSWORD") + "@tcp(" +
		os.Getenv("DB_HOST") + ":" +
		os.Getenv("DB_PORT") + ")/" +
		os.Getenv("DB_NAME") + "?charset=utf8mb4&parseTime=True&loc=Local"

	var err error
	DB, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("Error connecting to database: %v", err)
	}

	log.Println("Database connected successfully!")
}

// Close gracefully closes the database connection
func Close() {
	db, err := DB.DB()
	if err != nil {
		log.Fatalf("Error getting database instance: %v", err)
	}
	if err := db.Close(); err != nil {
		log.Fatalf("Error closing database: %v", err)
	}
}
