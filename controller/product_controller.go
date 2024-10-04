package controller

import (
	"backend/models"
	"backend/service"
	"net/http"
	"strconv"

	"log"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// ProductController handles HTTP requests related to products
type ProductController struct {
	productService     *service.ProductService
	TransactionService *service.TransactionService
	UserService        *service.UserService
}

// NewProductController creates a new ProductController instance
func NewProductController(productService *service.ProductService, transactionService *service.TransactionService, userService *service.UserService) *ProductController {
	return &ProductController{productService: productService, TransactionService: transactionService, UserService: userService}
}

// Helper function to parse UUID from the URL param
func parseUUIDParam(c *gin.Context, paramName string) (uuid.UUID, error) {
	idParam := c.Param(paramName)
	id, err := uuid.Parse(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid UUID format"})
		return uuid.Nil, err
	}
	return id, nil
}

// Create handles the creation of a new product
// @Summary      Create a new product
// @Description  Create a new product with the given details
// @Tags         Products
// @Accept       json
// @Produce      json
// @Param        product  body      models.ProductRequest  true  "Product data"
// @Success      201      {object}  models.ProductResponse
// @Router       /products [post]
func (controller *ProductController) Create(c *gin.Context) {
	var product models.ProductRequest
	if err := c.ShouldBindJSON(&product); err != nil {
		log.Printf("Create product: invalid input: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
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
	createdProduct, err := controller.productService.Create(&product, uid)
	if err != nil {
		log.Printf("Create product: service error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create product"})
		return
	}
	transaction := models.TransactionRequest{
		ItemID:      createdProduct.ID,                          // Use the ItemID from the request
		UserID:      uid,                                        // Use the UserID from the request
		Description: createdProduct.Description,                 // Use the Description from the request
		Action:      models.TransactionAction(models.Submitted), // Use the Action from the request (TransactionAction type)
		ImageURL:    product.ImageURL,                           // Use the ImageURL from the request
	}
	transactionCreated, _ := controller.TransactionService.AddTransaction(&transaction)
	user, _ := controller.UserService.GetDemographicInformation(uid.String())
	productResponse := models.ProductResponse{
		User:         *user,
		ID:           createdProduct.ID,
		Name:         createdProduct.Name,
		Description:  createdProduct.Description,
		Price:        createdProduct.Price,
		Category:     createdProduct.Category,
		SubCategory:  createdProduct.SubCategory,
		CreatedAt:    createdProduct.CreatedAt,
		Transactions: []models.Transaction{*transactionCreated},
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Product created successfully", "product": productResponse})
}

// ChangeStatus updates the status of a product by its ID.
// @Summary      Change Product Status
// @Description  Change the status of a product using its ID.
// @Tags         Products
// @Accept       json
// @Produce      json
// @Param        id      path      string                  true  "Product ID"
// @Param        status  body      models.ProductStatus    true  "New Status for the product"
// @Success      200     {object}  map[string]string       "Status updated successfully"
// @Failure      400     {object}  map[string]string       "Invalid input"
// @Failure      404     {object}  map[string]string       "Product not found"
// @Failure      500     {object}  map[string]string       "Failed to update status"
// @Router       /products/{id}/status [put]
func (controller *ProductController) ChangeStatus(c *gin.Context) {
	id := c.Param("id")
	var newStatus models.ProductStatus

	// Bind the new status
	if err := c.ShouldBindJSON(&newStatus); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input", "details": err.Error()})
		return
	}

	// Convert ID to UUID
	productID, err := uuid.Parse(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid product ID"})
		return
	}

	// Update the product's status
	err = controller.productService.UpdateStatus(productID, newStatus)
	if err != nil {
		if err == service.ErrProductNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Product not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update status", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Status updated successfully"})
}

// Delete handles the deletion of a product by its ID
// @Summary      Delete a product
// @Description  Delete a product by its UUID
// @Tags         Products
// @Param        id   path      string  true  "Product UUID"
// @Success      200   {object}  map[string]string @Router       /products/{id} [delete]
func (controller *ProductController) Delete(c *gin.Context) {
	id, err := parseUUIDParam(c, "id")
	if err != nil {
		return
	}

	if err := controller.productService.Delete(id); err != nil {
		log.Printf("Delete product: service error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete product"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Product deleted successfully"})
}

// GetOne retrieves a product by its ID
// @Summary Get a product by ID
// @Description Get a product by its unique ID
// @Param id path string true "Product ID"
// @Success 200 {object} models.ProductResponse
// @Failure 404 {object} gin.H{"error": "Product not found"}
// @Failure 500 {object} gin.H{"error": "Failed to retrieve user information"}
// @Router /products/{id} [get]
func (controller *ProductController) GetOne(c *gin.Context) {
	id, err := parseUUIDParam(c, "id")
	if err != nil {
		return
	}

	product, err := controller.productService.GetByID(id)
	if err != nil {
		log.Printf("GetOne product: service error: %v", err)
		c.JSON(http.StatusNotFound, gin.H{"error": "Product not found"})
		return
	}

	user, err := controller.UserService.GetDemographicInformation(product.UserID.String())
	if err != nil {
		log.Printf("GetOne product: failed to fetch user info: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve user information"})
		return
	}

	transactions, err := controller.TransactionService.GetByProductID(product.ID)
	if err != nil {
		log.Printf("GetOne product: failed to fetch transactions: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve transactions"})
		return
	}

	productResponse := models.ProductResponse{
		ID:           product.ID,
		User:         *user,
		Transactions: transactions,
		Name:         product.Name,
		Description:  product.Description,
		Price:        product.Price,
		SubCategory:  product.SubCategory,
		Category:     product.Category,
		CreatedAt:    product.CreatedAt,
	}

	c.JSON(http.StatusOK, gin.H{"product": productResponse})
}

// GetProductsByUserID retrieves products by user ID
// @Summary Get products by user ID
// @Description Get all products for a specific user
// @Success 200 {array} models.ProductResponse
// @Failure 401 {object} gin.H{"error": "User ID not found"}
// @Failure 500 {object} gin.H{"error": "Failed to retrieve user products"}
// @Router /products/user [get]
func (controller *ProductController) GetProductsByUserID(c *gin.Context) {
	localID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User ID not found"})
		return
	}

	userID, ok := localID.(uuid.UUID)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	products, err := controller.productService.GetProductsByUserID(userID)
	if err != nil {
		log.Printf("GetProductsByUserID: service error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve user products"})
		return
	}

	// Fetch user demographic information
	user, err := controller.UserService.GetDemographicInformation(userID.String())
	if err != nil {
		log.Printf("GetProductsByUserID: failed to fetch user info: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve user information"})
		return
	}

	// Prepare a response slice
	var productResponses []models.ProductResponse
	for _, product := range products {
		transactions, err := controller.TransactionService.GetByProductID(product.ID)
		if err != nil {
			log.Printf("GetProductsByUserID: failed to fetch transactions for product %s: %v", product.ID.String(), err)
			continue // Skip to the next product if there's an error
		}

		productResponses = append(productResponses, models.ProductResponse{
			ID:           product.ID,
			User:         *user,
			Transactions: transactions,
			Name:         product.Name,
			Description:  product.Description,
			Price:        product.Price,
			SubCategory:  product.SubCategory,
			Category:     product.Category,
			CreatedAt:    product.CreatedAt,
		})
	}

	c.JSON(http.StatusOK, gin.H{"products": productResponses})
}

// GetCollaborative retrieves products using a collaborative filtering approach with pagination
// @Summary Get collaborative recommendations
// @Description Retrieve products based on collaborative filtering
// @Param page query int false "Page number"
// @Success 200 {array} models.ProductResponse
// @Failure 500 {object} gin.H{"error": "Failed to retrieve collaborative products"}
// @Router /products/collaborative [get]
func (controller *ProductController) GetCollaborative(c *gin.Context) {
	localID, exists := c.Get("user_id")
	var userID uuid.UUID
	if exists {
		userID, _ = localID.(uuid.UUID)
	}

	// Get page number from query parameters, default to 1 if not provided
	page := 1
	if pageParam := c.Query("page"); pageParam != "" {
		if p, err := strconv.Atoi(pageParam); err == nil && p > 0 {
			page = p
		} else {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid page number"})
			return
		}
	}

	var products []models.Product
	var err error
	if exists {
		products, err = controller.productService.FetchCollaborativeRecommendations(userID, page)
		if err != nil {
			log.Printf("GetCollaborative: service error: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve collaborative products"})
			return
		}
	} else {
		// Return random products if userID is not available
		products, err = controller.productService.GetRandomProducts(userID)
		if err != nil {
			log.Printf("GetCollaborative: failed to fetch random products: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve random products"})
			return
		}
	}

	// Prepare product responses
	var productResponses []models.ProductResponse
	for _, product := range products {
		// Fetch user demographic information
		user, err := controller.UserService.GetDemographicInformation(product.UserID.String())
		if err != nil {
			log.Printf("GetCollaborative: failed to fetch user info for product %s: %v", product.ID.String(), err)
			continue // Skip to the next product if there's an error
		}

		transactions, err := controller.TransactionService.GetByProductID(product.ID)
		if err != nil {
			log.Printf("GetCollaborative: failed to fetch transactions for product %s: %v", product.ID.String(), err)
			continue // Skip to the next product if there's an error
		}

		productResponses = append(productResponses, models.ProductResponse{
			ID:           product.ID,
			User:         *user,
			Transactions: transactions,
			Name:         product.Name,
			Description:  product.Description,
			Price:        product.Price,
			SubCategory:  product.SubCategory,
			Category:     product.Category,
			CreatedAt:    product.CreatedAt,
		})
	}

	c.JSON(http.StatusOK, gin.H{"products": productResponses})
}

// GetContentBased retrieves products based on content-based filtering
// @Summary Get content-based recommendations
// @Description Retrieve products based on content-based filtering using an image URL
// @Param image_url query string true "Image URL"
// @Success 200 {array} models.ProductResponse
// @Failure 400 {object} gin.H{"error": "Invalid image URL"}
// @Failure 500 {object} gin.H{"error": "Failed to retrieve content-based products"}
// @Router /products/content-based [get]
func (controller *ProductController) GetContentBased(c *gin.Context) {
	// Retrieve image URL from query parameters
	imageURL := c.Query("image_url")
	if imageURL == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid image URL"})
		return
	}

	productIDs, err := controller.TransactionService.FetchContentBasedRecommendations(imageURL) // Updated method name
	if err != nil {
		log.Printf("GetContentBased: failed to fetch recommendations: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve content-based products"})
		return
	}

	products, err := controller.productService.GetProductsByIDs(productIDs)
	if err != nil {
		log.Printf("GetContentBased: service error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve products"})
		return
	}

	// Prepare product responses
	var productResponses []models.ProductResponse
	for _, product := range products {
		user, err := controller.UserService.GetDemographicInformation(product.UserID.String())
		if err != nil {
			log.Printf("GetContentBased: failed to fetch user info for product %s: %v", product.ID.String(), err)
			continue // Skip to the next product if there's an error
		}

		transactions, err := controller.TransactionService.GetByProductID(product.ID)
		if err != nil {
			log.Printf("GetContentBased: failed to fetch transactions for product %s: %v", product.ID.String(), err)
			continue // Skip to the next product if there's an error
		}

		productResponses = append(productResponses, models.ProductResponse{
			ID:           product.ID,
			User:         *user,
			Transactions: transactions,
			Name:         product.Name,
			Description:  product.Description,
			Price:        product.Price,
			SubCategory:  product.SubCategory,
			Category:     product.Category,
			CreatedAt:    product.CreatedAt,
		})
	}

	c.JSON(http.StatusOK, gin.H{"products": productResponses})
}
