package service

import (
	"backend/models"
	"backend/repository"

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

// HideTransaction hides a transaction by updating its hidden status
func (s *TransactionService) HideTransaction(id uuid.UUID) error {
	return s.transactionRepo.HideTransaction(id)
}

// GetByUserID retrieves transactions for a specific user
func (s *TransactionService) GetByUserID(userID uuid.UUID) ([]models.Transaction, error) {
	return s.transactionRepo.GetByUserID(userID)
}

// AddTransaction adds a transaction to an item
func (s *TransactionService) AddTransaction(transaction *models.Transaction) error {
	return s.transactionRepo.Create(transaction) // Assuming Create handles adding the transaction
}

// FetchContentBasedRecommendations retrieves products based on content filtering (mock implementation)
func (s *TransactionService) FetchContentBasedRecommendations(imageURL string) ([]uuid.UUID, error) {
	// TODO: Implement logic for fetching content-based recommendations based on the image URL
	return []uuid.UUID{}, nil
}
