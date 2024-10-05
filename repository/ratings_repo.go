// backend/repository/rating_repo.go
package repository

import (
	"backend/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// RatingRepository manages rating-related database interactions
type RatingRepository struct {
	db *gorm.DB
}

// NewRatingRepository initializes the repository with the database connection
func NewRatingRepository(db *gorm.DB) *RatingRepository {
	return &RatingRepository{db: db}
}

// Create adds a new rating to the database
func (repo *RatingRepository) Create(rating *models.Rating) error {
	return repo.db.Create(rating).Error
}

// Delete removes a rating by its ID
func (repo *RatingRepository) Delete(id uuid.UUID) error {
	return repo.db.Delete(&models.Rating{}, "id = ?", id).Error
}

// GetRatedProductsByUserId retrieves all rated products by a user's ID
func (repo *RatingRepository) GetRatedProductsByUserId(userID uuid.UUID) ([]models.Rating, error) {
	var ratings []models.Rating
	if err := repo.db.Where("user_id = ?", userID).Find(&ratings).Error; err != nil {
		return nil, err
	}
	return ratings, nil
}

// GetAverageRatingByProductId calculates the average rating and count for a product using GORM
func (r *RatingRepository) GetAverageRatingByProductId(productID uuid.UUID) (float64, int, error) {
	var result struct {
		Average float64
		Count   int
	}

	// Use GORM to calculate the average rating and count for the specified product
	err := r.db.Model(&models.Rating{}).
		Where("product_id = ?", productID).
		Select("AVG(score) as average, COUNT(*) as count").
		Scan(&result).Error

	if err != nil {
		return 0, 0, err
	}

	return result.Average, result.Count, nil
}

// GetPuanByUserIdItemId retrieves the rating given by a user for a specific product
func (r *RatingRepository) GetPuanByUserIdItemId(userID uuid.UUID, productID uuid.UUID) (int, error) {
	var rating models.Rating

	// Find the rating by user ID and product ID
	if err := r.db.Where("user_id = ? AND product_id = ?", userID, productID).First(&rating).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return 0, nil // Return 0 if no rating exists for this user and product
		}
		return 0, err // Return the error if something else went wrong
	}

	return int(rating.Score), nil
}
