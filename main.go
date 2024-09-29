package main

import (
	"backend/database"
	"backend/routes"
	"log"

	_ "backend/docs" // Import the generated Swagger docs

	"github.com/gofiber/fiber/v2"
	"github.com/joho/godotenv"
	fiberSwagger "github.com/swaggo/fiber-swagger" // Swagger handler
)

// @title Econova API
// @version 1.0
// @description This is a simple API for user registration
// @host localhost:3000
// @BasePath /
func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Fatalf("Error loading .env file")
	}

	app := fiber.New()

	// Connect to the database
	db := database.Connect()
	defer db.Close()

	// Swagger documentation route
	app.Use("/swagger/*", fiberSwagger.WrapHandler)

	// Setup routes
	routes.SetupRoutes(app, db)

	// Start the server
	if err := app.Listen(":3000"); err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
}
