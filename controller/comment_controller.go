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
	commentService service.CommentService
	userService    service.UserService
}

// NewCommentController creates a new CommentController instance
func NewCommentController(commentService service.CommentService, userService service.UserService) *CommentController {
	return &CommentController{
		commentService: commentService,
		userService:    userService,
	}
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

	// Get user_id from context (assumed to be set by middleware)
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
	id, err := uuid.Parse(idParam)
	if err != nil {
		log.Printf("Invalid comment UUID: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid UUID format"})
		return
	}

	// Get user_id from context
	userIDStr, exists := c.Get("user_id")
	if !exists {
		log.Println("User ID not found in request")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User authentication required"})
		return
	}

	// Parse user ID to uuid.UUID
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

	// Proceed to delete the comment
	if err := controller.commentService.Delete(id); err != nil {
		log.Printf("Error deleting comment: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete comment", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Comment deleted successfully"})
}

// GetByProductID retrieves all comments for a specific product, with user demographic information
// @Summary      Get comments by product ID
// @Description  Retrieves all comments for a specific product, with user demographic information
// @Tags         Comments
// @Accept       json
// @Produce      json
// @Param        product_id   path    string  true   "Product ID"
// @Success      200          {array} models.CommentResponse
// @Router       /comments/product/{product_id} [get]
func (controller *CommentController) GetByProductID(c *gin.Context) {
	productIDParam := c.Param("product_id")
	productID, err := uuid.Parse(productIDParam)
	if err != nil {
		log.Printf("Invalid product UUID: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid UUID format"})
		return
	}

	// Retrieve basic comments without User details from the service
	comments, err := controller.commentService.GetByProductID(productID)
	if err != nil {
		log.Printf("Error retrieving comments: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve comments", "details": err.Error()})
		return
	}

	// Create a slice to hold comments with full user details
	var commentsWithUserDetails []models.CommentResponse

	for _, comment := range comments {
		// Fetch demographic information for each user associated with a comment
		user, err := controller.userService.GetDemographicInformation(comment.UserID.String())
		if err != nil {
			log.Printf("Error fetching user demographic information: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve user information", "details": err.Error()})
			return
		}

		// Create a CommentResponse with the User information
		commentResponse := models.CommentResponse{
			ID:        comment.ID,
			User:      *user, // Populate the user information here
			ProductID: comment.ProductID,
			Content:   comment.Content,
			CreatedAt: comment.CreatedAt,
		}

		// Append to the list of comment responses
		commentsWithUserDetails = append(commentsWithUserDetails, commentResponse)
	}

	// Return the list of comments with user demographic information
	c.JSON(http.StatusOK, gin.H{"comments": commentsWithUserDetails})
}
