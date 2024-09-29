package routes

import (
	"database/sql"

	"github.com/gofiber/fiber/v2"
)

// SetupRoutes configures all the routes for the application
func SetupRoutes(app *fiber.App, db *sql.DB) {
	SetupPublicRoutes(app, db)
}
