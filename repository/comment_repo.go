package repository

import (
	"backend/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// CommentRepository manages database interactions for comments
type CommentRepository struct {
	db *gorm.DB
}

// NewCommentRepository initializes the repository with the database connection
func NewCommentRepository(db *gorm.DB) *CommentRepository {
	return &CommentRepository{db: db}
}

// Create adds a new comment to the database
func (repo *CommentRepository) Create(comment *models.Comment) error {
	return repo.db.Create(comment).Error
}

// Delete removes a comment by its ID
func (repo *CommentRepository) Delete(id uuid.UUID) error {
	return repo.db.Delete(&models.Comment{}, "id = ?", id).Error
}

// GetByProductID retrieves all comments for a specific product
func (repo *CommentRepository) GetByProductID(productID uuid.UUID) ([]models.Comment, error) {
	var comments []models.Comment
	if err := repo.db.Where("product_id = ?", productID).Find(&comments).Error; err != nil {
		return nil, err
	}
	return comments, nil
}

func (repo *CommentRepository) FindByID(id uuid.UUID) (*models.Comment, error) {
	var comment models.Comment
	// Search for the comment by its ID in the database
	err := repo.db.Where("id = ?", id).First(&comment).Error
	if err != nil {
		return nil, err // if not found or any other error, return error
	}
	return &comment, nil
}

// FindByUserAndProduct finds a comment by user and product ID
func (repo *CommentRepository) FindByUserAndProduct(userID uuid.UUID, productID uuid.UUID) (*models.Comment, error) {
	var comment models.Comment
	err := repo.db.Where("user_id = ? AND product_id = ?", userID, productID).First(&comment).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil // No comment found, return nil
		}
		return nil, err // Other errors
	}
	return &comment, nil // Return the found comment
}

// Update updates an existing comment
func (repo *CommentRepository) Update(comment *models.Comment) error {
	return repo.db.Save(comment).Error
}
