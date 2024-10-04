package service

import (
	"backend/models"
	"backend/repository"
	"time"

	"github.com/google/uuid"
)

// ProductService handles business logic for products
type ProductService struct {
	productRepo *repository.ProductRepository
}

// NewProductService creates a new instance of ProductService
func NewProductService(productRepo *repository.ProductRepository) *ProductService {
	return &ProductService{productRepo: productRepo}
}

// Create a new product
func (s *ProductService) Create(product *models.ProductRequest, userID uuid.UUID) (*models.Product, error) {
	p := &models.Product{ // Correctly initialize the Product struct
		ID:          uuid.New(),
		Name:        product.Name,
		Description: product.Description,
		Price:       product.Price,
		Category:    product.Category,
		Status:      models.StatusAvailable,
		SubCategory: product.SubCategory,
		CreatedAt:   time.Now().UTC(),
		UserID:      userID,
	}

	return p, s.productRepo.Create(p)
}

// Update an existing product
func (s *ProductService) Update(product *models.Product) error {
	return s.productRepo.Update(product)
}

// Delete a product by ID
func (s *ProductService) Delete(id uuid.UUID) error {
	return s.productRepo.Delete(id)
}

// GetByID retrieves a product by its ID
func (s *ProductService) GetByID(id uuid.UUID) (*models.Product, error) {
	return s.productRepo.GetByID(id)
}

// GetProductsByIDs retrieves products by their IDs
func (s *ProductService) GetProductsByIDs(ids []uuid.UUID) ([]models.Product, error) {
	return s.productRepo.GetProductsByIDs(ids)
}
func (s *ProductService) FetchCollaborativeRecommendations(userID uuid.UUID, page int) ([]models.Product, error) {
	// Here you would implement your collaborative filtering logic.
	// For example, you might fetch user purchase history, similar users, etc.

	// Placeholder logic:
	recommendedProductIDs := []uuid.UUID{} // This should be populated with actual logic

	products, err := s.productRepo.GetProductsByIDs(recommendedProductIDs) // Assuming you have a GetByIDs method
	if err != nil {
		return nil, err
	}
	return products, nil
}

// GetRandomProducts retrieves random products for a user
func (s *ProductService) GetRandomProducts(userID uuid.UUID) ([]models.Product, error) {
	return s.productRepo.GetRandomProducts()
}

// This is business logic that might filter or modify results before returning them.
func (s *ProductService) GetProductsByUserID(userID uuid.UUID) ([]models.Product, error) {
	products, err := s.productRepo.GetProductsByUserID(userID)
	if err != nil {
		return nil, err
	}

	// Business logic could be added here, e.g., filtering hidden products.
	return products, nil
}

// UpdateStatus updates the status of a product.
func (s *ProductService) UpdateStatus(productID uuid.UUID, status models.ProductStatus) error {
	product, err := s.productRepo.GetByID(productID)
	if err != nil {
		return ErrProductNotFound
	}

	product.Status = status
	return s.productRepo.Update(product)
}
