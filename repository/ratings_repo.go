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

// FindByUserAndProduct finds a rating by user and product
func (repo *RatingRepository) FindByUserAndProduct(userID uuid.UUID, productID uuid.UUID) (*models.Rating, error) {
	var rating models.Rating
	// Perform the query to find the rating by userID and productID
	err := repo.db.Where("user_id = ? AND product_id = ?", userID, productID).First(&rating).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil // No rating found, return nil
		}
		return nil, err // Some other error occurred
	}
	return &rating, nil // Return the found rating
}

// Update updates an existing rating in the database
func (repo *RatingRepository) Update(rating *models.Rating) error {
	// Perform the update operation
	if err := repo.db.Save(rating).Error; err != nil {
		return err
	}
	return nil
}
