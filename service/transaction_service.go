package service

import (
	"backend/models"
	"backend/repository"
	"time"

	"github.com/google/uuid"
)

// TransactionService handles business logic for transactions
type TransactionService struct {
	transactionRepo *repository.TransactionRepository
}

// NewTransactionService creates a new instance of TransactionService
func NewTransactionService(transactionRepo *repository.TransactionRepository) *TransactionService {
	return &TransactionService{transactionRepo: transactionRepo}
}

// Create a new transaction
func (s *TransactionService) Create(transaction *models.Transaction) error {
	return s.transactionRepo.Create(transaction)
}

// GetByUserID retrieves transactions for a specific user
func (s *TransactionService) GetByProductID(itemID uuid.UUID) ([]models.Transaction, error) {
	return s.transactionRepo.GetByProductID(itemID)
}

// AddTransaction adds a transaction to a product
func (s *TransactionService) AddTransaction(req *models.TransactionRequest) (*models.Transaction, error) {
	// Create a new Transaction object
	transaction := models.Transaction{
		ID:          uuid.New(),       // Generate a new UUID for the transaction
		ItemID:      req.ItemID,       // Use the ItemID from the request
		UserID:      req.UserID,       // Use the UserID from the request
		Description: req.Description,  // Use the Description from the request
		Action:      req.Action,       // Use the Action from the request (TransactionAction type)
		ImageURL:    req.ImageURL,     // Use the ImageURL from the request
		CreatedAt:   time.Now().UTC(), // Set CreatedAt to the current UTC time
	}

	// Save the transaction to the repository and return the transaction and any error
	err := s.transactionRepo.Create(&transaction)
	return &transaction, err
}

// FetchContentBasedRecommendations retrieves products based on content filtering (mock implementation)
func (s *TransactionService) FetchContentBasedRecommendations(imageURL string) ([]uuid.UUID, error) {
	// TODO: Implement actual logic for fetching content-based recommendations based on the image URL
	return []uuid.UUID{}, nil
}

// / GetProductIDsByImageURLs retrieves product IDs associated with a list of image URLs
func (s *TransactionService) GetProductIDsByImageURLs(imageURLs []string) ([]uuid.UUID, error) {
	// Step 1: Fetch transactions related to the image URLs
	transactions, err := s.transactionRepo.GetByImageURLs(imageURLs)
	if err != nil {
		return nil, err
	}

	// Step 2: Extract ItemIDs (Product IDs) from transactions
	var productIDs []uuid.UUID
	for _, transaction := range transactions {
		productIDs = append(productIDs, transaction.ItemID)
	}

	return productIDs, nil
}
