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

// GetByProductID retrieves transactions for a specific item ID, ordered by created timestamp.
func (r *TransactionRepository) GetByProductID(itemID uuid.UUID) ([]models.Transaction, error) {
	var transactions []models.Transaction
	if err := r.db.Where("item_id = ?", itemID).
		Order("created_at DESC"). // Order by created_at in descending order
		Find(&transactions).Error; err != nil {
		return nil, err
	}
	return transactions, nil
}

func (r *TransactionRepository) GetByImageURLs(imageURLs []string) ([]models.Transaction, error) {
	var transactions []models.Transaction
	err := r.db.Where("image_url IN ?", imageURLs).Find(&transactions).Error
	if err != nil {
		return nil, err
	}
	return transactions, nil
}
