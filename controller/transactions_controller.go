package controller

import (
	"backend/models"
	"backend/service"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// TransactionController handles HTTP requests related to transactions
type TransactionController struct {
	transactionService *service.TransactionService
}

// NewTransactionController creates a new TransactionController instance
func NewTransactionController(transactionService *service.TransactionService) *TransactionController {
	return &TransactionController{transactionService: transactionService}
}

// addTransactionToItem adds a transaction to an item
// @Summary      Add transaction to item
// @Description  Adds a transaction (submitted, revitalized, or sold) to an item
// @Tags         Transactions
// @Accept       json
// @Produce      json
// @Param        item_id      path      string  true   "Item ID"
// @Param        user_id      path      string  true   "User ID"
// @Param        body         body      models.Transaction  true   "Transaction details"
// @Success      201          {object}  models.Transaction
// @Router       /transactions/item/{item_id}/user/{user_id} [post]
func (controller *TransactionController) AddTransactionToItem(c *gin.Context) {
	var transaction models.Transaction
	if err := c.ShouldBindJSON(&transaction); err != nil {
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

	userID, err := uuid.Parse(c.Param("user_id"))
	if err != nil {
		log.Printf("Invalid user UUID: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user UUID format"})
		return
	}

	transaction.ItemID = itemID
	transaction.UserID = userID

	// Add transaction
	if err := controller.transactionService.AddTransaction(&transaction); err != nil {
		log.Printf("Error adding transaction: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add transaction", "details": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Transaction added successfully", "transaction": transaction})
}

// hideTransaction hides a transaction based on its ID
// @Summary      Hide transaction
// @Description  Hides a transaction by updating the hidden flag to true
// @Tags         Transactions
// @Accept       json
// @Produce      json
// @Param        id      path   string  true   "Transaction ID"
// @Router       /transactions/{id}/hide [patch]
func (controller *TransactionController) HideTransaction(c *gin.Context) {
	idParam := c.Param("id")
	transactionID, err := uuid.Parse(idParam)
	if err != nil {
		log.Printf("Invalid transaction UUID: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid transaction UUID format"})
		return
	}

	// Hide transaction
	if err := controller.transactionService.HideTransaction(transactionID); err != nil {
		log.Printf("Error hiding transaction: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hide transaction", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Transaction hidden successfully"})
}
