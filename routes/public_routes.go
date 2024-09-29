package routes

import (
	"backend/controllers"
	"backend/repository"
	"database/sql"

	"github.com/gofiber/fiber/v2"
)

// SetupPublicRoutes sets up the public routes
func SetupPublicRoutes(app *fiber.App, db *sql.DB) {
	repo := repository.NewUserRepository(db)
	userController := controllers.NewUserController(repo)
	homeController := controllers.NewHomeController()

	// Define public routes
	app.Get("/", homeController.Home)          // Home route
	app.Post("/signup", userController.SignUp) // Signup route
	app.Post("/login", userController.Login)   // Login route
}
