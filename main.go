package main

import (
	"backend/database"
	"backend/routes"
	"log"
	"os"

	_ "backend/docs" // Import the generated Swagger docs

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	swagFiles "github.com/swaggo/files"        // Correct import for Swagger files
	ginSwagger "github.com/swaggo/gin-swagger" // Import the gin-swagger package
)

// @title Econova API
// @version 1.0
// @description This is a simple API for user registration
// @host localhost:3000
// @BasePath /
func main() {
	// Load environment variables
	loadEnv()

	// Connect to the database
	database.Connect()     // Call the Connect function
	defer database.Close() // Ensure the database connection is closed when the function exits

	// Initialize Gin router
	router := gin.Default()

	// Setup CORS middleware
	setupCORS(router)

	// Setup Swagger documentation
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swagFiles.Handler)) // Use swagFiles.Handler

	// Setup routes
	routes.SetupRoutes(router, database.DB) // Pass the GORM DB instance to the routes

	// Start the server
	if err := router.Run(":3000"); err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
}

// loadEnv loads environment variables from a .env file
func loadEnv() {
	if err := godotenv.Load(); err != nil {
		log.Panicf("Error loading .env file: %v", err)
	}
}

// setupCORS configures CORS middleware for the Gin router
// setupCORS configures CORS middleware for the Gin router
func setupCORS(router *gin.Engine) {
	router.Use(cors.New(cors.Config{
		AllowOrigins: []string{os.Getenv("FE_PORT")}, // Change this to your frontend's URL or use "*" for all
		AllowMethods: []string{"GET,POST,PUT,DELETE,OPTIONS"},
		AllowHeaders: []string{"Origin, Content-Type, Accept, Authorization"},
	}))
}
