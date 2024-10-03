package controller

import (
	"backend/models"
	"backend/service"
	"errors"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// CommentsController handles HTTP requests related to comments
type CommentsController struct {
	commentsService *service.CommentsService
}

// NewCommentsController creates a new CommentsController instance
func NewCommentsController(commentsService *service.CommentsService) *CommentsController {
	return &CommentsController{commentsService: commentsService}
}

// Create handles the creation of a new comment
// @Summary      Create a Comment
// @Description  Create a new comment for a product by a user
// @Tags         Comments
// @Accept       json
// @Produce      json
// @Param        comment  body  models.Comment  true  "Comment data"
// @Success      201      {object} models.CommentResponse
// @Router       /comments [post]
func (controller *CommentsController) Create(c *gin.Context) {
	var comment models.Comment

	// Fetch user ID from context (assuming user is authenticated and user ID is stored in context)
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	// Parse the incoming request
	if err := c.ShouldBindJSON(&comment); err != nil {
		log.Printf("Error parsing JSON request: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}

	// Assign the authenticated user's ID to the comment
	comment.UserID = userID.(uuid.UUID)

	// Handle comment creation logic
	if err := controller.commentsService.Create(&comment); err != nil {
		if errors.Is(err, service.ErrInvalidComment) {
			log.Printf("Invalid comment data: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid comment data"})
		} else if errors.Is(err, service.ErrDatabase) {
			log.Printf("Database error: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		} else {
			log.Printf("Unexpected error: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create comment"})
		}
		return
	}

	c.JSON(http.StatusCreated, comment)
}

// HideComment handles hiding a comment
// @Summary      Hide a Comment
// @Description  Hide a specific comment by ID
// @Tags         Comments
// @Accept       json
// @Produce      json
// @Param        comment_id  path  string  true  "Comment ID"
// @Success      200         {object} models.CommentResponse
// @Router       /comments/{comment_id}/hide [post]
func (controller *CommentsController) HideComment(c *gin.Context) {
	commentID := c.Param("comment_id")

	// Validate the comment ID format (UUID in this case)
	if _, err := uuid.Parse(commentID); err != nil {
		log.Printf("Invalid comment ID format: %v", commentID)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid comment ID format"})
		return
	}

	// Fetch user ID from context
	userID, exists := c.Get("user_id")
	if !exists {
		log.Printf("User not authenticated")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	// Hide comment logic
	if err := controller.commentsService.HideComment(userID.(uuid.UUID), commentID); err != nil {
		if errors.Is(err, service.ErrCommentNotFound) {
			log.Printf("Comment not found: %v", commentID)
			c.JSON(http.StatusNotFound, gin.H{"error": "Comment not found"})
		} else if errors.Is(err, service.ErrUnauthorized) {
			log.Printf("Unauthorized action for comment: %v", commentID)
			c.JSON(http.StatusForbidden, gin.H{"error": "You do not have permission to hide this comment"})
		} else {
			log.Printf("Error hiding comment: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hide comment"})
		}
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Comment hidden successfully"})
}

// GetCommentsByUserId retrieves comments by user ID
// @Summary      Get Comments by User ID
// @Description  Retrieve all comments made by a specific user
// @Tags         Comments
// @Accept       json
// @Produce      json
// @Param        user_id  path  string  true  "User ID"
// @Success      200      {array}  models.Comment
// @Router       /comments/user/{user_id} [get]
func (controller *CommentsController) GetCommentsByUserId(c *gin.Context) {
	localID, exists := c.Get("user_id") // Get the user ID from locals
	var userID uuid.UUID
	if exists {
		userID, _ = localID.(uuid.UUID)
	}

	userIDParam := c.Param("user_id")

	// Validate the user ID format (UUID in this case)
	if _, err := uuid.Parse(userIDParam); err != nil {
		log.Printf("Invalid user ID format: %v", userIDParam)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID format"})
		return
	}

	// Fetch comments
	comments, err := controller.commentsService.GetCommentsByUserId(userIDParam)
	if err != nil {
		if errors.Is(err, service.ErrUserNotFound) {
			log.Printf("User not found: %v", userIDParam)
			c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		} else {
			log.Printf("Error fetching comments: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch comments"})
		}
		return
	}

	// Filter comments based on visibility
	var filteredComments []models.Comment
	for _, comment := range comments {
		if !comment.Hidden || (comment.Hidden && comment.UserID == userID) { // Adjust according to your Comment model
			filteredComments = append(filteredComments, comment)
		}
	}

	c.JSON(http.StatusOK, filteredComments)
}

// GetCommentsByItemId retrieves comments by product/item ID
// @Summary      Get Comments by Product ID
// @Description  Retrieve all comments made on a specific product
// @Tags         Comments
// @Accept       json
// @Produce      json
// @Param        product_id  path  string  true  "Product ID"
// @Success      200         {array}  models.CommentResponse
// @Router       /comments/product/{product_id} [get]
func (controller *CommentsController) GetCommentsByItemId(c *gin.Context) {
	localID, exists := c.Get("user_id") // Get the user ID from locals
	var userID uuid.UUID
	if exists {
		userID, _ = localID.(uuid.UUID)
	}

	productID := c.Param("product_id")

	// Validate the product ID format (UUID in this case)
	if _, err := uuid.Parse(productID); err != nil {
		log.Printf("Invalid product ID format: %v", productID)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid product ID format"})
		return
	}

	// Fetch comments
	comments, err := controller.commentsService.GetCommentsByItemId(productID)
	if err != nil {
		if errors.Is(err, service.ErrProductNotFound) {
			log.Printf("Product not found: %v", productID)
			c.JSON(http.StatusNotFound, gin.H{"error": "Product not found"})
		} else {
			log.Printf("Error fetching comments: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch comments"})
		}
		return
	}

	// Filter comments based on visibility
	var filteredComments []models.Comment
	for _, comment := range comments {
		if !comment.Hidden || (comment.Hidden && comment.UserID == userID) { // Adjust according to your Comment model
			filteredComments = append(filteredComments, comment)
		}
	}

	c.JSON(http.StatusOK, filteredComments)
}
