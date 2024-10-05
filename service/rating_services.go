// backend/service/rating_service.go
package service

import (
	"backend/models"
	"backend/repository"

	"github.com/google/uuid"
)

// RatingService handles the business logic for ratings
type RatingService struct {
	ratingRepo *repository.RatingRepository
}

// NewRatingService creates a new RatingService instance
func NewRatingService(ratingRepo *repository.RatingRepository) *RatingService {
	return &RatingService{ratingRepo: ratingRepo}
}

// Create adds a new rating using the rating repository
func (service *RatingService) Create(rating *models.Rating) error {
	return service.ratingRepo.Create(rating)
}

// Delete removes a rating by its ID
func (service *RatingService) Delete(id uuid.UUID) error {
	return service.ratingRepo.Delete(id)
}

// GetRatedProductsByUserId retrieves all rated products by a user's ID
func (service *RatingService) GetRatedProductsByUserId(userID uuid.UUID) ([]models.Rating, error) {
	return service.ratingRepo.GetRatedProductsByUserId(userID)
}
func (service *RatingService) GetPuanByUserIdItemId(userID uuid.UUID, itemId uuid.UUID) (int, error) {
	puan, err := service.ratingRepo.GetPuanByUserIdItemId(userID, itemId)
	if err != nil {
		return 0, nil
	}
	return puan, nil

}

// GetAverageRatingByProductId retrieves the average rating and the count of ratings for a product
func (s *RatingService) GetAverageRatingByProductId(productID uuid.UUID) (float64, int, error) {
	// Logic to calculate the average rating and count goes here
	average, count, err := s.ratingRepo.GetAverageRatingByProductId(productID)
	if err != nil {
		return 0, 0, err
	}
	return average, count, nil
}
