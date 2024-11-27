package service

import (
	"backend/models"
	"backend/repository"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/joho/godotenv"
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
// Update updates the product and returns an error if the operation fails
func (service *ProductService) Update(product *models.Product) error {
	fmt.Println(product)
	err := service.productRepo.Update(product)
	if err != nil {
		return fmt.Errorf("failed to update product: %w", err)
	}
	return nil
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
func (s *ProductService) FetchCollaborativeRecommendations(userID string) ([]models.Product, error) {
	err := godotenv.Load()
	if err != nil {
		return nil, fmt.Errorf("failed to load .env file: %v", err)
	}

	// Get Flask server URL from environment variable
	url := fmt.Sprintf("%s?user_id=%s", os.Getenv("FLASK_SERVER_URL2"), userID)
	//url := fmt.Sprintf("http://localhost:5001/recommendations?user_id=%s", userID)

	// Make the HTTP GET request to fetch recommendations
	response, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close() // Ensure the response body is closed

	// Check if the response status is OK
	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("error fetching recommendations: status %d", response.StatusCode)
	}

	// Parse the JSON response
	var recommendations struct {
		UserID          string             `json:"user_id"`
		Recommendations map[string]float64 `json:"recommendations"`
	}

	if err := json.NewDecoder(response.Body).Decode(&recommendations); err != nil {
		return nil, err
	}

	// Extract product IDs from recommendations
	recommendedProductIDs := make([]uuid.UUID, 0, len(recommendations.Recommendations))
	for productIDStr := range recommendations.Recommendations {
		productID, err := uuid.Parse(productIDStr)
		if err != nil {
			return nil, fmt.Errorf("invalid product ID: %s", productIDStr)
		}
		recommendedProductIDs = append(recommendedProductIDs, productID)
	}

	// If the number of recommended product IDs is less than the threshold, fetch random products
	if len(recommendedProductIDs) < 10 {
		additionalProducts, err := s.productRepo.GetRandomProducts()
		if err != nil {
			return nil, err
		}

		// Combine the recommended product IDs with the random products
		for _, product := range additionalProducts {
			recommendedProductIDs = append(recommendedProductIDs, product.ID)
		}
	}

	// Retrieve product details based on the recommended product IDs
	products, err := s.productRepo.GetProductsByIDs(recommendedProductIDs)
	if err != nil {
		return nil, err
	}

	// Limit the number of products to a maximum of 10
	if len(products) > 10 {
		products = products[:10]
	}

	return products, nil
}

// FetchItemBasedRecommendations fetches recommendations for an item based on collaborative filtering
func (s *ProductService) FetchItemBasedRecommendations(productID string) ([]models.Product, error) {
	err := godotenv.Load()
	if err != nil {
		return nil, fmt.Errorf("failed to load .env file: %v", err)
	}

	// Get Flask server URL from environment variable (Item-based URL)
	url := fmt.Sprintf("%s?product_id=%s", os.Getenv("FLASK_SERVER_URL2"), productID)
	// url := fmt.Sprintf("http://localhost:5001/recommendations?product_id=%s", productID)

	// Make the HTTP GET request to fetch recommendations
	response, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close() // Ensure the response body is closed

	// Check if the response status is OK
	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("error fetching recommendations: status %d", response.StatusCode)
	}

	// Parse the JSON response for item-based recommendations
	var recommendations struct {
		ProductID    string             `json:"product_id"`
		SimilarItems map[string]float64 `json:"similar_items"`
	}

	if err := json.NewDecoder(response.Body).Decode(&recommendations); err != nil {
		return nil, err
	}

	// Extract product IDs from recommendations
	recommendedProductIDs := make([]uuid.UUID, 0, len(recommendations.SimilarItems))
	for productIDStr := range recommendations.SimilarItems {
		productID, err := uuid.Parse(productIDStr)
		if err != nil {
			return nil, fmt.Errorf("invalid product ID: %s", productIDStr)
		}
		recommendedProductIDs = append(recommendedProductIDs, productID)
	}

	// If the number of recommended product IDs is less than the threshold, fetch random products
	if len(recommendedProductIDs) < 10 {
		additionalProducts, err := s.productRepo.GetRandomProducts()
		if err != nil {
			return nil, err
		}

		// Combine the recommended product IDs with the random products
		for _, product := range additionalProducts {
			recommendedProductIDs = append(recommendedProductIDs, product.ID)
		}
	}

	// Retrieve product details based on the recommended product IDs
	products, err := s.productRepo.GetProductsByIDs(recommendedProductIDs)
	if err != nil {
		return nil, err
	}

	// Limit the number of products to a maximum of 10
	if len(products) > 10 {
		products = products[:10]
	}

	return products, nil
}

// GetRandomProducts retrieves random products for a user
func (s *ProductService) GetRandomProducts() ([]models.Product, error) {
	return s.productRepo.GetRandomProducts()
}

// GetProductsByUserID retrieves products for a specific user by their UUID with pagination
func (s *ProductService) GetProductsByUserID(userID uuid.UUID, count, page int) ([]models.Product, error) {
	// Calculate offset based on page and count
	offset := (page - 1) * count

	// Call the repository method with pagination parameters
	products, err := s.productRepo.GetProductsByUserID(userID, count, offset)
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

// GetRestoredProducts retrieves products with the status "restored"// GetProductsByStatusPaginated fetches products by the specified status with pagination
func (s *ProductService) GetProductsByStatusPaginated(status string, limit int, offset int) ([]models.Product, error) {
	// Call repository method to get products by status with pagination
	return s.productRepo.GetByStatusPaginated(status, limit, offset)
}

func (s *ProductService) GetRandomProductsPaginated(count int, offset int) ([]models.Product, error) {
	// Call the repository function to get random products with pagination
	products, err := s.productRepo.GetRandomProductsPaginated(count, offset)
	if err != nil {
		return nil, err
	}

	return products, nil
}
