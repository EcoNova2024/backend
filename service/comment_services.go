package service

import (
	"backend/models"
	"backend/repository"
	"errors"

	"github.com/google/uuid"
)

// CommentService defines the interface for comment services
type CommentService interface {
	Create(commentData *models.AddComment, userID string) (*models.Comment, error)
	Delete(id uuid.UUID) error
	GetByProductID(productID uuid.UUID) ([]models.Comment, error)
	Update(id uuid.UUID, content string) (*models.Comment, error)
	GetByID(id uuid.UUID) (*models.Comment, error)
}

// commentService is the concrete implementation of CommentService
type commentService struct {
	repo *repository.CommentRepository
}

// NewCommentService creates a new instance of CommentService
func NewCommentService(repo *repository.CommentRepository) CommentService {
	return &commentService{repo: repo}
}

// Create adds a new comment
func (s *commentService) Create(commentData *models.AddComment, userID string) (*models.Comment, error) {
	if commentData.Content == "" {
		return nil, errors.New("content cannot be empty")
	}

	productID, err := uuid.Parse(commentData.ProductID)
	if err != nil {
		return nil, errors.New("invalid product ID")
	}

	comment := &models.Comment{
		UserID:    uuid.MustParse(userID),
		ProductID: productID,
		Content:   commentData.Content,
	}

	if err := s.repo.Create(comment); err != nil {
		return nil, err
	}

	return comment, nil
}

// Delete removes a comment by ID
func (s *commentService) Delete(id uuid.UUID) error {
	return s.repo.Delete(id)
}

// GetByProductID fetches all comments for a given product
func (s *commentService) GetByProductID(productID uuid.UUID) ([]models.Comment, error) {
	return s.repo.GetByProductID(productID)
}

// GetByID retrieves a comment by its ID from the repository
func (s *commentService) GetByID(id uuid.UUID) (*models.Comment, error) {
	// Call the repository to find the comment by its ID
	comment, err := s.repo.FindByID(id)
	if err != nil {
		return nil, err // return the error if comment is not found or any other error occurs
	}
	return comment, nil
}

// Update updates the content of an existing comment
func (s *commentService) Update(id uuid.UUID, content string) (*models.Comment, error) {
	if content == "" {
		return nil, errors.New("content cannot be empty")
	}

	comment, err := s.repo.FindByUserAndProduct(id, id) // Example UUID check
	if err != nil {
		return nil, err
	}

	comment.Content = content
	if err := s.repo.Update(comment); err != nil {
		return nil, err
	}

	return comment, nil
}
