package service

import (
	"backend/models"
	"backend/repository"
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/joho/godotenv"
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
func (s *TransactionService) FetchContentBasedRecommendations(imageFilename string) ([]uuid.UUID, error) {
	err := godotenv.Load()
	if err != nil {
		return nil, fmt.Errorf("failed to load .env file: %v", err)
	}

	// Get Flask server URL from environment variable
	url := os.Getenv("FLASK_SERVER_URL")

	// Create the request payload
	payload := map[string]string{"filename": imageFilename}
	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal JSON: %v", err)
	}

	// Send the POST request to the Python Flask application
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonPayload))
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %v", err)
	}
	defer resp.Body.Close()

	// Check for successful response
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("received non-200 response: %d", resp.StatusCode)
	}

	// Decode the response directly into a slice of maps
	var similarImages []map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&similarImages); err != nil {
		return nil, fmt.Errorf("failed to decode response: %v", err)
	}

	// Extract image URLs for fetching item IDs
	var imageURLs []string
	for _, img := range similarImages {
		if name, ok := img["name"].(string); ok {
			imageURLs = append(imageURLs, name) // Extract the image name
		}
	}

	// Fetch transactions by image URLs
	fetchedTransactions, err := s.transactionRepo.GetByImageURLs(imageURLs)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch transactions: %v", err)
	}

	// Extract item IDs from fetched transactions
	var itemIDs []uuid.UUID
	for _, t := range fetchedTransactions {
		itemIDs = append(itemIDs, t.ItemID)
	}

	return itemIDs, nil
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
