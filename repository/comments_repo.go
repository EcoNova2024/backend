// backend/repository/comments_repository.go
package repository

import (
	"backend/models"
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Define repository-specific errors
var ErrRecordNotFound = errors.New("record not found")

// CommentsRepository handles the database operations for comments
type CommentsRepository struct {
	db *gorm.DB
}

// NewCommentsRepository creates a new instance of CommentsRepository
func NewCommentsRepository(db *gorm.DB) *CommentsRepository {
	return &CommentsRepository{db: db}
}

// Create inserts a new comment into the database
func (repo *CommentsRepository) Create(comment *models.Comment) error {
	if err := repo.db.Create(comment).Error; err != nil {
		return err
	}
	return nil
}

// GetByID retrieves a comment by its ID
func (repo *CommentsRepository) GetByID(id uuid.UUID) (*models.Comment, error) {
	var comment models.Comment
	if err := repo.db.First(&comment, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrRecordNotFound
		}
		return nil, err
	}
	return &comment, nil
}

// Update updates an existing comment
func (repo *CommentsRepository) Update(comment *models.Comment) error {
	if err := repo.db.Save(comment).Error; err != nil {
		return err
	}
	return nil
}

// GetByUserID retrieves comments by a user ID
func (repo *CommentsRepository) GetByUserID(userID uuid.UUID) ([]models.Comment, error) {
	var comments []models.Comment
	if err := repo.db.Where("user_id = ?", userID).Find(&comments).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrRecordNotFound
		}
		return nil, err
	}
	return comments, nil
}

// GetByProductID retrieves comments by product/item ID
func (repo *CommentsRepository) GetByProductID(productID uuid.UUID) ([]models.Comment, error) {
	var comments []models.Comment
	if err := repo.db.Where("product_id = ?", productID).Find(&comments).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrRecordNotFound
		}
		return nil, err
	}
	return comments, nil
}
