package main

import (
	"backend/database"
	"backend/routes"
	"log"

	_ "backend/docs" // Import the generated Swagger docs

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/swagger"
	"github.com/joho/godotenv"
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
	// Connect to the database
	db := database.Connect()
	defer db.Close()
	app := fiber.New()
	app.Use(cors.New(cors.Config{
		AllowOrigins: "*", // Change this to your frontend's URL
		AllowMethods: "GET,POST,PUT,DELETE,OPTIONS",
		AllowHeaders: "Origin, Content-Type, Accept, Authorization",
	}))
	app.Use("/swagger/*", swagger.HandlerDefault)
	// Setup routes
	routes.SetupRoutes(app, db)

	// Start the server
	if err := app.Listen(":3000"); err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
}
