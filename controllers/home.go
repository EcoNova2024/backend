package controllers

import "github.com/gofiber/fiber/v2"

// HomeController handles home route
type HomeController struct{}

// NewHomeController creates a new instance of HomeController
func NewHomeController() *HomeController {
	return &HomeController{}
}

// Home displays the welcome message
// @Summary      Home
// @Description  Display welcome message
// @Tags         Home
// @Produce      json
// @Router       / [get]
func (hc *HomeController) Home(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{
		"message": "Welcome to the Econova API!",
	})
}
