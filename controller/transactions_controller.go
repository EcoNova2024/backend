package controller

import (
	"backend/models"
	"backend/service"
	"fmt"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// TransactionController handles HTTP requests related to transactions
type TransactionController struct {
	transactionService *service.TransactionService
	productService     *service.ProductService
}

// NewTransactionController creates a new TransactionController instance
func NewTransactionController(transactionService *service.TransactionService, productService *service.ProductService) *TransactionController {
	return &TransactionController{transactionService: transactionService, productService: productService}
}

// AddTransactionToItem adds a transaction to an item
// @Summary      Add transaction to item
// @Description  Adds a transaction (submitted, revitalized, or sold) to an item
// @Tags         Transactions
// @Accept       json
// @Produce      json
// @Param        item_id      path      string  true   "Item ID"
// @Param        body         body      models.AddTransactionRequest  true   "Transaction details"
// @Success      201          {object}  models.Transaction
// @Router       /transactions/{item_id} [post]
func (controller *TransactionController) AddTransactionToItem(c *gin.Context) {
	var transactionReq models.AddTransactionRequest
	if err := c.ShouldBindJSON(&transactionReq); err != nil {
		log.Printf("Error binding JSON: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input", "details": err.Error()})
		return
	}

	// Validate UUID fields
	itemID, err := uuid.Parse(c.Param("item_id"))
	if err != nil {
		log.Printf("Invalid item UUID: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid item UUID format"})
		return
	}

	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User ID not found in context"})
		return
	}

	// Convert userID to UUID
	uid, err := uuid.Parse(userID.(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID format"})
		return
	}

	// Retrieve product details by ID
	product, err := controller.productService.GetByID(itemID)
	if err != nil {
		log.Printf("Error retrieving product: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve product", "details": err.Error()})
		return
	}
	if product == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Product not found"})
		return
	}

	// Prepare transaction data
	transaction := models.TransactionRequest{
		ItemID:      itemID,
		UserID:      uid,
		Description: transactionReq.Description, // Use the Description from the request
		Action:      transactionReq.Action,      // Use the Action from the request (TransactionAction type)
		ImageURL:    transactionReq.ImageURL,    // Use the ImageURL from the request
	}

	// Add the transaction
	t, err := controller.transactionService.AddTransaction(&transaction)
	if err != nil {
		log.Printf("Error adding transaction: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add transaction", "details": err.Error()})
		return
	}
	product.UserID = transaction.UserID
	product.Price = transactionReq.Price
	if transactionReq.Action == "revitalized" {
		product.Status = "restored"
	} else {
		if transaction.Action == "submited" {
			product.Status = "available"

		} else {
			product.Status = "sold"
		}

	}
	err = controller.productService.Update(product)
	if err != nil {
		log.Printf("Error updating product status: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update product status", "details": err.Error()})
		return
	}
	fmt.Println(product.Status)

	c.JSON(http.StatusCreated, gin.H{"message": "Transaction added successfully", "transaction": t})
}
