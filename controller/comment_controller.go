package controller

import (
	"backend/models"
	"backend/service"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// CommentController handles HTTP requests related to comments
type CommentController struct {
	commentService service.CommentService // Use the interface, not the concrete type pointer
}

// NewCommentController creates a new CommentController instance
func NewCommentController(commentService service.CommentService) *CommentController {
	return &CommentController{commentService: commentService}
}

// Create handles the creation of a new comment
// @Summary      Create a new comment
// @Description  Creates a new comment for a product by a user
// @Tags         Comments
// @Accept       json
// @Produce      json
// @Param        body  body   models.AddComment  true  "Comment details"
// @Success      201   {object}  models.Comment
// @Router       /comments [post]
func (controller *CommentController) Create(c *gin.Context) {
	var addComment models.AddComment
	if err := c.ShouldBindJSON(&addComment); err != nil {
		log.Printf("Error binding JSON: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input", "details": err.Error()})
		return
	}

	// Get user_id from locals (assuming it's set in middleware)
	userID, exists := c.Get("user_id")
	if !exists {
		log.Println("User ID not found in request")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User authentication required"})
		return
	}

	// Call the service to create the comment
	comment, err := controller.commentService.Create(&addComment, userID.(string))
	if err != nil {
		log.Printf("Error creating comment: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create comment", "details": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Comment created successfully", "comment": comment})
}

// Delete handles the deletion of a comment by its ID
// @Summary      Delete a comment
// @Description  Deletes a comment by its ID
// @Tags         Comments
// @Accept       json
// @Produce      json
// @Param        id   path   string  true   "Comment ID"
// @Router       /comments/{id} [delete]
func (controller *CommentController) Delete(c *gin.Context) {
	// Extract the comment ID from the URL parameters
	idParam := c.Param("id")
	id, err := uuid.Parse(idParam) // Parse the comment ID
	if err != nil {
		log.Printf("Invalid comment UUID: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid UUID format"})
		return
	}

	// Get user_id from JWT middleware (which is likely stored as a string)
	userIDStr, exists := c.Get("user_id")
	if !exists {
		log.Println("User ID not found in request")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User authentication required"})
		return
	}

	// Convert userID string from context into uuid.UUID
	userID, err := uuid.Parse(userIDStr.(string))
	if err != nil {
		log.Printf("Invalid user UUID: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user UUID format"})
		return
	}

	// Retrieve the comment from the service
	comment, err := controller.commentService.GetByID(id)
	if err != nil {
		log.Printf("Error retrieving comment: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve comment", "details": err.Error()})
		return
	}

	// Check if the user is the author of the comment
	if comment.UserID != userID {
		log.Println("Unauthorized attempt to delete comment")
		c.JSON(http.StatusForbidden, gin.H{"error": "You are not authorized to delete this comment"})
		return
	}

	// Proceed to delete the comment if the user is authorized
	if err := controller.commentService.Delete(id); err != nil {
		log.Printf("Error deleting comment: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete comment", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Comment deleted successfully"})
}

// GetByProductID handles retrieving all comments by product ID
// @Summary      Get comments by product ID
// @Description  Retrieves all comments for a specific product
// @Tags         Comments
// @Accept       json
// @Produce      json
// @Param        product_id   path    string  true   "Product ID"
// @Success      200          {array} models.Comment
// @Router       /comments/product/{product_id} [get]
func (controller *CommentController) GetByProductID(c *gin.Context) {
	productIDParam := c.Param("product_id")
	productID, err := uuid.Parse(productIDParam)
	if err != nil {
		log.Printf("Invalid product UUID: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid UUID format"})
		return
	}

	comments, err := controller.commentService.GetByProductID(productID)
	if err != nil {
		log.Printf("Error retrieving comments: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve comments", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"comments": comments})
}
