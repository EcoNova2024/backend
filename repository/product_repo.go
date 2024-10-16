package repository

import (
	"github.com/EcoNova2024/backend/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ProductRepository handles database operations for products
type ProductRepository struct {
	db *gorm.DB
}

// NewProductRepository creates a new instance of ProductRepository
func NewProductRepository(db *gorm.DB) *ProductRepository {
	return &ProductRepository{db: db}
}

// Create inserts a new product into the database
func (r *ProductRepository) Create(product *models.Product) error {
	return r.db.Create(product).Error
}

// Update updates an existing product in the database
func (repo *ProductRepository) Update(product *models.Product) error {
	// Using GORM's Save method to update the entire product struct
	if err := repo.db.Save(product).Error; err != nil {
		return err
	}
	return nil
}

// Delete removes a product from the database by ID
func (r *ProductRepository) Delete(id uuid.UUID) error {
	return r.db.Delete(&models.Product{}, id).Error
}

// GetByID retrieves a product by its ID
func (r *ProductRepository) GetByID(id uuid.UUID) (*models.Product, error) {
	var product models.Product
	if err := r.db.First(&product, id).Error; err != nil {
		return nil, err
	}
	return &product, nil
}

// GetProductsByIDs retrieves multiple products by their IDs
func (r *ProductRepository) GetProductsByIDs(ids []uuid.UUID) ([]models.Product, error) {
	var products []models.Product
	if err := r.db.Find(&products, ids).Error; err != nil {
		return nil, err
	}
	return products, nil
}

// GetRandomProducts retrieves a specified number of random products
func (r *ProductRepository) GetRandomProducts() ([]models.Product, error) {
	var products []models.Product
	// Adjust the limit as needed
	if err := r.db.Order("RAND()").Limit(10).Find(&products).Error; err != nil {
		return nil, err
	}
	return products, nil
}

// GetProductsByUserID retrieves products for a specific user by their UUID
func (r *ProductRepository) GetProductsByUserID(userID uuid.UUID) ([]models.Product, error) {
	var products []models.Product
	if err := r.db.Where("user_id = ?", userID).Find(&products).Error; err != nil {
		return nil, err
	}
	return products, nil
}

// GetByStatus retrieves 10 random products by their status
func (r *ProductRepository) GetByStatus(status string) ([]models.Product, error) {
	var products []models.Product
	if err := r.db.Where("status = ?", status).Order("RAND()").Limit(10).Find(&products).Error; err != nil {
		return nil, err
	}
	return products, nil
}
