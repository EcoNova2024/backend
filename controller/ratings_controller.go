package controller

import (
	"backend/models"
	"backend/service"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// RatingController handles HTTP requests related to ratings
type RatingController struct {
	ratingService *service.RatingService
}

// NewRatingController creates a new RatingController instance
func NewRatingController(ratingService *service.RatingService) *RatingController {
	return &RatingController{ratingService: ratingService}
}

// Create handles the creation of a new rating
// @Summary      Create a new rating
// @Description  Creates a new rating for a product by a user
// @Tags         Ratings
// @Accept       json
// @Produce      json
// @Param        user_id    path      string  true   "User ID"
// @Param        product_id path      string  true   "Product ID"
// @Param        body       body      models.Rating  true   "Rating details"
// @Success      201        {object}  models.Rating
// @Router       /ratings/{user_id}/{product_id} [post]
func (controller *RatingController) Create(c *gin.Context) {
	var rating models.Rating
	if err := c.ShouldBindJSON(&rating); err != nil {
		log.Printf("Error binding JSON: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input", "details": err.Error()})
		return
	}

	// Validate UUID fields
	userID, err := uuid.Parse(c.Param("user_id"))
	if err != nil {
		log.Printf("Invalid user UUID: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user UUID format"})
		return
	}

	productID, err := uuid.Parse(c.Param("product_id"))
	if err != nil {
		log.Printf("Invalid product UUID: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid product UUID format"})
		return
	}

	rating.UserID = userID
	rating.ProductID = productID

	// Create rating
	if err := controller.ratingService.Create(&rating); err != nil {
		log.Printf("Error creating rating: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create rating", "details": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Rating created successfully", "rating": rating})
}

// Delete handles the deletion of a rating by its ID
// @Summary      Delete a rating
// @Description  Deletes a rating by its ID
// @Tags         Ratings
// @Accept       json
// @Produce      json
// @Param        id   path   string  true   "Rating ID"
// @Router       /ratings/{id} [delete]
func (controller *RatingController) Delete(c *gin.Context) {
	idParam := c.Param("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		log.Printf("Invalid rating UUID: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid UUID format"})
		return
	}

	if err := controller.ratingService.Delete(id); err != nil {
		log.Printf("Error deleting rating: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete rating", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Rating deleted successfully"})
}

// GetRatedProductsByUserId retrieves all rated products by a user's ID
// @Summary      Get rated products by user ID
// @Description  Retrieves all products rated by a specific user
// @Tags         Ratings
// @Accept       json
// @Produce      json
// @Param        user_id   path    string  true   "User ID"
// @Success      200       {array} models.Rating
// @Router       /ratings/user/{user_id} [get]
func (controller *RatingController) GetRatedProductsByUserId(c *gin.Context) {
	userIDParam := c.Param("user_id")
	userID, err := uuid.Parse(userIDParam)
	if err != nil {
		log.Printf("Invalid user UUID: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user UUID format"})
		return
	}

	ratings, err := controller.ratingService.GetRatedProductsByUserId(userID)
	if err != nil {
		log.Printf("Error retrieving rated products: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve rated products", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"ratings": ratings})
}

// GetAverageRatingByProductId retrieves the average rating and the count of ratings for a product
// @Summary      Get average rating and count by product ID
// @Description  Retrieves the average rating and the total number of ratings for a specific product
// @Tags         Ratings
// @Accept       json
// @Produce      json
// @Param        product_id  path    string  true   "Product ID"
// @Success      200         {object}  map[string]interface{}
// @Router       /ratings/product/{product_id}/average [get]
func (controller *RatingController) GetAverageRatingByProductId(c *gin.Context) {
	productIDParam := c.Param("product_id")
	productID, err := uuid.Parse(productIDParam)
	if err != nil {
		log.Printf("Invalid product UUID: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid product UUID format"})
		return
	}

	// Call the service to get the average rating and count
	average, count, err := controller.ratingService.GetAverageRatingByProductId(productID)
	if err != nil {
		log.Printf("Error retrieving average rating: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve average rating", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"average_rating": average,
		"rating_count":   count,
	})
}
