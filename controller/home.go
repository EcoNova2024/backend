package controller

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// HomeController handles HTTP requests for the home endpoint
type HomeController struct{}

// NewHomeController creates a new HomeController instance
func NewHomeController() *HomeController {
	return &HomeController{}
}

// Index handles requests to the home endpoint
func (controller *HomeController) Index(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"message": "Welcome to the API!",
		"status":  "Running",
	})
}
