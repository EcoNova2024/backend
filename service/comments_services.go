// backend/service/comments_service.go
package service

import (
	"backend/models"
	"backend/repository"
	"errors"
	"log"

	"github.com/google/uuid"
)

// Define specific errors for handling within the service layer
var (
	ErrInvalidComment  = errors.New("invalid comment data")
	ErrCommentNotFound = errors.New("comment not found")
	ErrUserNotFound    = errors.New("user not found")
	ErrProductNotFound = errors.New("product not found")
	ErrDatabase        = errors.New("database error")
	ErrUnauthorized    = errors.New("unauthorized request")
)

// CommentsService handles the business logic for comments
type CommentsService struct {
	commentsRepo *repository.CommentsRepository
}

// NewCommentsService creates a new instance of CommentsService
func NewCommentsService(commentsRepo *repository.CommentsRepository) *CommentsService {
	return &CommentsService{commentsRepo: commentsRepo}
}

// Create adds a new comment to the database
func (service *CommentsService) Create(comment *models.Comment) error {
	// Example of validation: Ensure comment content is not empty
	if comment.Content == "" || comment.UserID == uuid.Nil || comment.ProductID == uuid.Nil {
		log.Printf("Invalid comment data: %v", comment)
		return ErrInvalidComment
	}

	// Interact with the repository to save the comment
	if err := service.commentsRepo.Create(comment); err != nil {
		log.Printf("Database error: %v", err)
		return ErrDatabase
	}
	return nil
}

// HideComment handles the logic to hide a comment if the user owns it
func (service *CommentsService) HideComment(userID uuid.UUID, commentID string) error {
	// Parse the comment ID and ensure it's valid
	id, err := uuid.Parse(commentID)
	if err != nil {
		return ErrInvalidComment
	}

	// Check if the comment exists
	comment, err := service.commentsRepo.GetByID(id)
	if err != nil {
		if errors.Is(err, repository.ErrRecordNotFound) {
			log.Printf("Comment not found: %v", commentID)
			return ErrCommentNotFound
		}
		log.Printf("Database error: %v", err)
		return ErrDatabase
	}

	// Check if the comment belongs to the user
	if comment.UserID != userID {
		log.Printf("Unauthorized attempt to hide comment: %v by user: %v", commentID, userID)
		return ErrUnauthorized
	}

	// Hide the comment (assuming there's a Hidden field to update)
	comment.Hidden = true
	if err := service.commentsRepo.Update(comment); err != nil {
		log.Printf("Failed to hide comment: %v", err)
		return ErrDatabase
	}
	return nil
}

// GetCommentsByUserId retrieves comments by user ID
func (service *CommentsService) GetCommentsByUserId(userID string) ([]models.Comment, error) {
	id, err := uuid.Parse(userID)
	if err != nil {
		return nil, ErrInvalidComment
	}

	comments, err := service.commentsRepo.GetByUserID(id)
	if err != nil {
		if errors.Is(err, repository.ErrRecordNotFound) {
			log.Printf("User not found: %v", userID)
			return nil, ErrUserNotFound
		}
		log.Printf("Database error: %v", err)
		return nil, ErrDatabase
	}

	return comments, nil
}

// GetCommentsByItemId retrieves comments by product/item ID
func (service *CommentsService) GetCommentsByItemId(productID string) ([]models.Comment, error) {
	id, err := uuid.Parse(productID)
	if err != nil {
		return nil, ErrInvalidComment
	}

	comments, err := service.commentsRepo.GetByProductID(id)
	if err != nil {
		if errors.Is(err, repository.ErrRecordNotFound) {
			log.Printf("Product not found: %v", productID)
			return nil, ErrProductNotFound
		}
		log.Printf("Database error: %v", err)
		return nil, ErrDatabase
	}

	return comments, nil
}
