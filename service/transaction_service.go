package service

import (
	"backend/models"
	"backend/repository"
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
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

func (service *TransactionService) handleTransactionImage(transaction *models.Transaction) error {
	// Check if the Transaction has an image URL
	if transaction.ImageURL != "" {
		// Construct the S3 object key for the Transaction's image
		imageKey := fmt.Sprintf("images/%s", transaction.ImageURL)

		// Use the GetImage utility to get the pre-signed URL
		preSignedURL, err := GetImage(imageKey)
		if err != nil {
			return fmt.Errorf("failed to retrieve image URL: %v", err)
		}

		// Replace the ImageURL with the pre-signed URL
		transaction.ImageURL = preSignedURL
	} else {
		// If no image URL is provided, set it to an empty string
		transaction.ImageURL = ""
	}
	return nil
}

// GetByUserID retrieves transactions for a specific user
func (s *TransactionService) GetByProductID(itemID uuid.UUID) ([]models.Transaction, error) {
	// Retrieve transactions for the specific item ID
	transactions, err := s.transactionRepo.GetByProductID(itemID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch transactions: %v", err)
	}

	// Handle the image URL for each transaction
	for i := range transactions {
		err := s.handleTransactionImage(&transactions[i])
		if err != nil {
			return nil, fmt.Errorf("failed to handle image URL for transaction: %v", err)
		}
	}

	return transactions, nil
}

// AddTransaction adds a transaction to a product
func (s *TransactionService) AddTransaction(req *models.TransactionRequest) (*models.Transaction, error) {
	// Log the start of the AddTransaction process
	log.Printf("Adding transaction for ItemID: %s, UserID: %s", req.ItemID, req.UserID)

	// Create a new Transaction object
	transaction := models.Transaction{
		ID:          uuid.New(),       // Generate a new UUID for the transaction
		ItemID:      req.ItemID,       // Use the ItemID from the request
		UserID:      req.UserID,       // Use the UserID from the request
		Description: req.Description,  // Use the Description from the request
		Action:      req.Action,       // Use the Action from the request (TransactionAction type)
		CreatedAt:   time.Now().UTC(), // Set CreatedAt to the current UTC time
	}

	// Log transaction creation
	log.Printf("Created new transaction with ID: %s", transaction.ID)

	// Handle the image URL for the transaction (upload image and set pre-signed URL if necessary)
	err := s.handleTransactionPutImage(&transaction, req)
	if err != nil {
		log.Printf("Error handling image for transaction ID %s: %v", transaction.ID, err)
		return nil, fmt.Errorf("failed to handle image URL: %v", err)
	}

	// Save the transaction to the repository
	err = s.transactionRepo.Create(&transaction)
	if err != nil {
		log.Printf("Error saving transaction ID %s to the repository: %v", transaction.ID, err)
		return nil, fmt.Errorf("failed to save transaction: %v", err)
	}

	// Log success
	log.Printf("Successfully added transaction with ID: %s", transaction.ID)

	// Return the transaction
	return &transaction, nil
}

// handleTransactionPutImage decodes and uploads an image if present in req.ImageData
func (s *TransactionService) handleTransactionPutImage(transaction *models.Transaction, req *models.TransactionRequest) error {
	// Log the image handling process
	log.Printf("Handling image for transaction ID: %s", transaction.ID)

	// Check if ImageData contains base64-encoded image data
	if req.ImageData != "" {
		// Log that image data is provided
		log.Printf("Image data found for transaction ID: %s", transaction.ID)

		// Decode the base64-encoded image data
		imageData, err := base64.StdEncoding.DecodeString(req.ImageData)
		if err != nil {
			log.Printf("Error decoding base64 image data for transaction ID %s: %v", transaction.ID, err)
			return fmt.Errorf("failed to decode image data: %v", err)
		}

		// Generate a unique key for the image based on the transaction ID
		imageKey := fmt.Sprintf("%s.jpg", transaction.ID.String())

		// Log image key generation
		log.Printf("Generated image key for transaction ID %s: %s", transaction.ID, imageKey)

		// Upload the image using the S3 service's PutImage method and get a pre-signed URL
		imageURL, err := PutImage("images/"+imageKey, imageData)
		if err != nil {
			log.Printf("Error uploading image for transaction ID %s: %v", transaction.ID, err)
			return fmt.Errorf("failed to upload image: %v", err)
		}

		// Set the transaction's ImageURL to the pre-signed URL returned by PutImage
		transaction.ImageURL = imageKey
		log.Printf("Successfully uploaded image for transaction ID %s, URL: %s", transaction.ID, imageURL)
	}

	return nil
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
