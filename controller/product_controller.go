package controller

import (
	"backend/models"
	"backend/service"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// ProductController handles HTTP requests related to products
type ProductController struct {
	productService     *service.ProductService
	TransactionService *service.TransactionService
	UserService        *service.UserService
	RatingService      *service.RatingService
}

// NewProductController creates a new ProductController instance
func NewProductController(productService *service.ProductService, transactionService *service.TransactionService, userService *service.UserService, ratingService *service.RatingService) *ProductController {
	return &ProductController{
		productService:     productService,
		TransactionService: transactionService,
		UserService:        userService,
		RatingService:      ratingService,
	}
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
		ItemID:      createdProduct.ID,
		UserID:      uid,
		Description: createdProduct.Description,
		Action:      models.TransactionAction(models.Submitted),
		ImageURL:    product.ImageURL,
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

// GetOne retrieves a product by its ID
// @Summary Get a product by ID
// @Description Get a product by its unique ID
// @Param id query string true "Product ID"
// @Success 200 {object} models.ProductResponse
// @Failure 404 {object} gin.H{"error": "Product not found"}
// @Failure 500 {object} gin.H{"error": "Failed to retrieve user information"}
// @Router /products [get]
func (controller *ProductController) GetOne(c *gin.Context) {
	id := c.Query("id") // Retrieve the product ID from the query parameter
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing product ID"})
		return
	}

	productID, err := uuid.Parse(id) // Parse the string ID to UUID
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid product ID format"})
		return
	}

	product, err := controller.productService.GetByID(productID)
	if err != nil {
		log.Printf("GetOne product: service error: %v", err)
		c.JSON(http.StatusNotFound, gin.H{"error": "Product not found"})
		return
	}

	productResponse, err := controller.populateAdditionalProductData(product)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve additional product data"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"product": productResponse})
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
		productResponse, err := controller.populateAdditionalProductData(&product)
		if err != nil {
			log.Printf("GetProductsByUserID: failed to fetch additional data for product %s: %v", product.ID.String(), err)
			continue // Skip to the next product if there's an error
		}

		productResponses = append(productResponses, productResponse)
	}

	c.JSON(http.StatusOK, gin.H{"products": productResponses})
}

// GetProductsByUserID retrieves products by user ID
// @Summary Get products by user ID
// @Description Get all products for a specific user
// @Param user_id query string true "User ID"
// @Success 200 {array} models.ProductResponse
// @Failure 400 {object} gin.H{"error": "Invalid user ID format"}
// @Failure 404 {object} gin.H{"error": "User not found"}
// @Failure 500 {object} gin.H{"error": "Failed to retrieve user products"}
// @Router /products/user [get]
func (controller *ProductController) GetProductsByUserID(c *gin.Context) {
	userIDStr := c.Query("user_id") // Retrieve the user ID from the query parameter
	if userIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing user ID"})
		return
	}

	userID, err := uuid.Parse(userIDStr) // Parse the string user ID to UUID
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID format"})
		return
	}

	products, err := controller.productService.GetProductsByUserID(userID)
	if err != nil {
		log.Printf("GetProductsByUserID: service error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve user products"})
		return
	}

	var productResponses []models.ProductResponse
	for _, product := range products {
		productResponse, err := controller.populateAdditionalProductData(&product)
		if err != nil {
			log.Printf("GetProductsByUserID: failed to fetch additional data for product %s: %v", product.ID.String(), err)
			continue // Skip to the next product if there's an error
		}

		productResponses = append(productResponses, productResponse)
	}

	c.JSON(http.StatusOK, gin.H{"products": productResponses})
}

// GetCollaborative retrieves products using a collaborative filtering approach
// @Summary Get collaborative recommendations
// @Description Retrieve products based on collaborative filtering
// @Success 200 {array} models.ProductResponse
// @Failure 500 {object} gin.H{"error": "Failed to retrieve collaborative products"}
// @Router /products/collaborative [get]
func (controller *ProductController) GetCollaborative(c *gin.Context) {
	localID, exists := c.Get("user_id")
	var userID uuid.UUID
	if exists {
		userID, _ = localID.(uuid.UUID)
	}

	var products []models.Product
	var err error
	if exists {
		products, err = controller.productService.FetchCollaborativeRecommendations(userID)
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

	var productResponses []models.ProductResponse
	for _, product := range products {
		productResponse, err := controller.populateAdditionalProductData(&product)
		if err != nil {
			log.Printf("GetCollaborative: failed to fetch additional data for product %s: %v", product.ID.String(), err)
			continue // Skip to the next product if there's an error
		}

		productResponses = append(productResponses, productResponse)
	}

	c.JSON(http.StatusOK, gin.H{"products": productResponses})
}

func (controller *ProductController) populateAdditionalProductData(product *models.Product) (models.ProductResponse, error) {
	var productRes models.ProductResponse
	transactions, err := controller.TransactionService.GetByProductID(product.ID)
	if err != nil {
		return productRes, err
	}
	productRes.Transactions = transactions

	// Fetch average rating and rating count
	averageRating, ratingCount, err := controller.RatingService.GetAverageRatingByProductId(product.ID)
	if err != nil {
		return productRes, err
	}
	user, _ := controller.UserService.GetDemographicInformation(product.UserID.String())
	productRes.User = *user
	productRes.Rating, _ = controller.RatingService.GetPuanByUserIdItemId(product.UserID, product.ID)
	productRes.RatingAverage = averageRating
	productRes.RatingCount = ratingCount

	return productRes, nil
}
