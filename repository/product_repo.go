package repository

import (
	"backend/models"

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

// GetProductsByUserID retrieves products for a specific user by their UUID with pagination
func (r *ProductRepository) GetProductsByUserID(userID uuid.UUID, count, offset int) ([]models.Product, error) {
	var products []models.Product
	// Add limit and offset for pagination
	if err := r.db.Where("user_id = ?", userID).Limit(count).Offset(offset).Find(&products).Error; err != nil {
		return nil, err
	}
	return products, nil
}

// GetByStatus retrieves 10 random products by their status// GetByStatusPaginated retrieves products by any given status with pagination
func (repo *ProductRepository) GetByStatusPaginated(status string, limit int, offset int) ([]models.Product, error) {
	var products []models.Product

	// Query to fetch products by status with pagination
	err := repo.db.
		Where("status = ?", status). // Filter by the specified status
		Order("created_at DESC").    // Order by CreatedAt in descending order
		Limit(limit).                // Limit results for pagination
		Offset(offset).              // Start from the specified offset
		Find(&products).             // Execute query
		Error

	if err != nil {
		return nil, err
	}

	return products, nil
}

func (repo *ProductRepository) GetRandomProductsPaginated(count int, offset int) ([]models.Product, error) {
	var products []models.Product

	// Query to fetch random products with pagination, ordered by CreatedAt
	err := repo.db.
		Order("created_at DESC"). // Ensure they're ordered by CreatedAt descending
		Limit(count).             // Limit the results to the count
		Offset(offset).           // Start from the specified offset
		Find(&products).          // Perform the query and load results into products
		Error

	if err != nil {
		return nil, err
	}

	return products, nil
}
