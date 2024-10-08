// backend/service/rating_service.go
package service

import (
	"backend/models"
	"backend/repository"
	"errors"
	"log"
	"time"

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

// Create adds or updates a rating using the rating repository
func (service *RatingService) Create(addRating *models.AddRating, userID string) (*models.Rating, error) {
	// Validate user ID
	parsedUserID, err := uuid.Parse(userID)
	if err != nil {
		log.Printf("Invalid user UUID: %v", err)
		return nil, errors.New("invalid user UUID format")
	}

	// Validate ProductID
	productID, err := uuid.Parse(addRating.ProductID)
	if err != nil {
		log.Printf("Invalid product UUID: %v", err)
		return nil, errors.New("invalid product UUID format")
	}

	// Check if a rating by this user for this product already exists
	existingRating, err := service.ratingRepo.FindByUserAndProduct(parsedUserID, productID)
	if err != nil {
		log.Printf("Error finding existing rating: %v", err)
		return nil, errors.New("failed to check existing rating")
	}

	if existingRating != nil {
		// If rating exists, update the score and updated_at timestamp
		existingRating.Score = addRating.Score
		existingRating.CreatedAt = time.Now().UTC()

		if err := service.ratingRepo.Update(existingRating); err != nil {
			log.Printf("Error updating rating: %v", err)
			return nil, errors.New("failed to update rating")
		}

		return existingRating, nil
	}

	// Create a new rating if none exists
	rating := &models.Rating{
		UserID:    parsedUserID,
		ProductID: productID,
		Score:     addRating.Score,
		CreatedAt: time.Now().UTC(),
	}

	if err := service.ratingRepo.Create(rating); err != nil {
		log.Printf("Error creating rating: %v", err)
		return nil, errors.New("failed to create rating")
	}

	return rating, nil
}

// Delete removes a rating by its ID
func (service *RatingService) Delete(id uuid.UUID) error {
	if err := service.ratingRepo.Delete(id); err != nil {
		log.Printf("Error deleting rating with ID %s: %v", id, err)
		return errors.New("failed to delete rating")
	}
	return nil
}

// GetRatedProductsByUserId retrieves all rated products by a user's ID
func (service *RatingService) GetRatedProductsByUserId(userID uuid.UUID) ([]models.Rating, error) {
	ratings, err := service.ratingRepo.GetRatedProductsByUserId(userID)
	if err != nil {
		log.Printf("Error retrieving ratings for user ID %s: %v", userID, err)
		return nil, errors.New("failed to retrieve user ratings")
	}
	return ratings, nil
}

// GetPuanByUserIdItemId retrieves the score (puan) for a specific user and item
func (service *RatingService) GetPuanByUserIdItemId(userID uuid.UUID, itemId uuid.UUID) (int, error) {
	puan, _ := service.ratingRepo.FindByUserAndProduct(userID, itemId)

	if puan == nil {
		return 0, errors.New("no rating found")
	}
	return int(puan.Score), nil

}

// GetAverageRatingByProductId retrieves the average rating and the count of ratings for a product
func (service *RatingService) GetAverageRatingByProductId(productID uuid.UUID) (float64, int, error) {
	average, count, err := service.ratingRepo.GetAverageRatingByProductId(productID)
	if err != nil {
		log.Printf("Error retrieving average rating for product ID %s: %v", productID, err)
		return 0, 0, errors.New("failed to retrieve average rating for product")
	}
	return average, count, nil
}

func (service *RatingService) GetRatedProductIDsByUserID(userID string) ([]string, error) {
	// Delegate to repository to get product IDs
	return service.ratingRepo.GetRatedItemsByUserID(userID)
}
