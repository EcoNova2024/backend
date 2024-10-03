package repository

import (
	"backend/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// TransactionRepository handles database operations for transactions
type TransactionRepository struct {
	db *gorm.DB
}

// NewTransactionRepository creates a new instance of TransactionRepository
func NewTransactionRepository(db *gorm.DB) *TransactionRepository {
	return &TransactionRepository{db: db}
}

// Create inserts a new transaction into the database
func (r *TransactionRepository) Create(transaction *models.Transaction) error {
	return r.db.Create(transaction).Error
}

// HideTransaction updates a transaction to set its hidden status
func (r *TransactionRepository) HideTransaction(id uuid.UUID) error {
	return r.db.Model(&models.Transaction{}).Where("id = ?", id).Update("hidden", true).Error
}

// GetByUserID retrieves transactions for a specific user
func (r *TransactionRepository) GetByUserID(userID uuid.UUID) ([]models.Transaction, error) {
	var transactions []models.Transaction
	if err := r.db.Where("user_id = ?", userID).Find(&transactions).Error; err != nil {
		return nil, err
	}
	return transactions, nil
}
