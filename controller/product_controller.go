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
// @Router /products/content-based [get]
func (controller *ProductController) GetContentBased(c *gin.Context) {
	// Retrieve image URL from query parameters
	imageURL := c.Query("image_url")
	if imageURL == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid image URL"})
		return
	}

	// Try to fetch content-based recommendations
	productIDs, err := controller.TransactionService.FetchContentBasedRecommendations(imageURL)
	if err != nil {
		log.Printf("GetContentBased: failed to fetch recommendations: %v", err)

		// Fallback to get random products if content-based fetching fails
		products, err := controller.productService.GetRandomProducts()
		if err != nil {
			log.Printf("GetContentBased: failed to fetch random products: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve random products"})
			return
		}

		// Prepare product responses for random products
		var randomProductResponses []models.ProductResponse
		for _, product := range products {
			productResponse, err := controller.populateAdditionalProductData(&product)
			if err != nil {
				log.Printf("GetProductsByUserID: failed to fetch additional data for random product %s: %v", product.ID.String(), err)
				continue // Skip to the next product if there's an error
			}

			randomProductResponses = append(randomProductResponses, productResponse)
		}

		c.JSON(http.StatusOK, gin.H{"products": randomProductResponses})
		return
	}

	// If content-based recommendations are successful, proceed as usual
	products, err := controller.productService.GetProductsByIDs(productIDs)
	if err != nil {
		log.Printf("GetContentBased: service error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve products"})
		return
	}

	// Prepare product responses for fetched products
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
// @Router /products/collaborative [get]
func (controller *ProductController) GetCollaborative(c *gin.Context) {
	// Attempt to retrieve the user ID from the context
	localID, exists := c.Get("user_id")
	if !exists {
		log.Println("User ID not found in request")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User authentication required"})
		return
	}

	// Assert the user ID type to string or uuid.UUID as necessary
	userID, ok := localID.(string) // or uuid.UUID based on your implementation
	if !ok {
		log.Println("Invalid user ID type")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user ID"})
		return
	}

	// Fetch collaborative recommendations
	products, err := controller.productService.FetchCollaborativeRecommendations(userID)
	if err != nil {
		log.Printf("GetCollaborative: service error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve collaborative products"})
		return
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

	// Return successful response with populated product data
	c.JSON(http.StatusOK, gin.H{"products": productResponses})
}

// GetRandomProducts retrieves random products when the user is not logged in
// @Summary Get random products
// @Description Retrieve random products for unauthenticated users
// @Success 200 {array} models.ProductResponse
// @Router /products/random [get]
func (controller *ProductController) GetRandomProducts(c *gin.Context) {
	// Fetch random products from the product service
	products, err := controller.productService.GetRandomProducts()
	if err != nil {
		log.Printf("GetRandomProducts: failed to fetch random products: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve random products"})
		return
	}

	var productResponses []models.ProductResponse
	for _, product := range products {
		productResponse, err := controller.populateAdditionalProductData(&product)
		if err != nil {
			log.Printf("GetRandomProducts: failed to fetch additional data for product %s: %v", product.ID.String(), err)
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

	// Fetch average rating and rating count
	averageRating, ratingCount, err := controller.RatingService.GetAverageRatingByProductId(product.ID)
	if err != nil {
		return productRes, err
	}
	user, _ := controller.UserService.GetDemographicInformation(product.UserID.String())

	UserRating, _ := controller.RatingService.GetPuanByUserIdItemId(product.UserID, product.ID)

	productRes = models.ProductResponse{
		User:          *user,
		ID:            product.ID,
		Name:          product.Name,
		Description:   product.Description,
		Price:         product.Price,
		Category:      product.Category,
		SubCategory:   product.SubCategory,
		RatingCount:   ratingCount,
		RatingAverage: averageRating,
		Rating:        UserRating,
		CreatedAt:     product.CreatedAt,
		Status:        product.Status,
		Transactions:  transactions,
	}
	return productRes, nil
}

// GetRestoredProducts retrieves products with the status "restored"
// @Summary Get restored products
// @Description Retrieve products with the status "restored"
// @Success 200 {array} models.ProductResponse
// @Router /products/restored [get]
func (controller *ProductController) GetRestoredProducts(c *gin.Context) {
	// Fetch restored products from the product service
	products, err := controller.productService.GetRestoredProducts()
	if err != nil {
		log.Printf("GetRestoredProducts: failed to fetch restored products: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve restored products"})
		return
	}

	var productResponses []models.ProductResponse
	for _, product := range products {
		productResponse, err := controller.populateAdditionalProductData(&product)
		if err != nil {
			log.Printf("GetRestoredProducts: failed to fetch additional data for product %s: %v", product.ID.String(), err)
			continue // Skip to the next product if there's an error
		}

		productResponses = append(productResponses, productResponse)
	}

	c.JSON(http.StatusOK, gin.H{"products": productResponses})
}
