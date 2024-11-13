package controller

import (
	"backend/models"
	"backend/service"
	"log"
	"net/http"
	"strconv"

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

// @Summary      Create a new product with image
// @Description  Create a new product with the given details, including Base64-encoded image data
// @Tags         Products
// @Accept       json
// @Produce      json
// @Param        product  body      models.ProductRequest  true  "Product data"
// @Param        image_data  body   string                  true  "Base64 encoded image data"
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
		ImageData:   product.ImageData,
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
// @Tags         Products
// @Description Get a product by its unique ID
// @Param id query string true "Product ID"
// @Success 200 {object} models.DetailedProductResponse
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
	detailedProductResponse, err := controller.populateAdditionalTransactionData(&productResponse)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve additional Transaction data"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"product": detailedProductResponse})
}

// GetContentBased retrieves products based on content-based filtering
// @Summary Get content-based recommendations
// @Tags         Products
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

// GetProductsByUserID retrieves products by user ID with pagination
// @Summary Get products by user ID with pagination
// @Tags Products
// @Description Get all products for a specific user with pagination support
// @Param user_id query string true "User ID"
// @Param count   query int    true "Number of products per page"
// @Param page    query int    true "Page number"
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

	// Get pagination parameters
	count, err := strconv.Atoi(c.DefaultQuery("count", "10")) // Default to 10 if not provided
	if err != nil || count <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid count value"})
		return
	}

	page, err := strconv.Atoi(c.DefaultQuery("page", "1")) // Default to 1 if not provided
	if err != nil || page <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid page value"})
		return
	}

	// Call the service to get products with pagination
	products, err := controller.productService.GetProductsByUserID(userID, count, page)
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

	// Return the paginated products
	c.JSON(http.StatusOK, gin.H{"products": productResponses})
}

// GetCollaborative retrieves products using a collaborative filtering approach
// @Summary Get collaborative recommendations
// @Tags         Products
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
// @Tags         Products
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

// GetProductsByStatus retrieves products by a specified status with pagination
// @Summary Get products by status
// @Tags         Products
// @Description Retrieve products by the specified status with pagination
// @Param        status  query string true  "Product status (e.g., restored, active, archived)"
// @Param        limit   query int    false "Number of products per page"
// @Param        page    query int    false "Page number"
// @Success 200  {array} models.ProductResponse
// @Router /products/status [get]
func (controller *ProductController) GetProductsByStatus(c *gin.Context) {
	// Retrieve the status parameter from the query string
	status := c.Query("status")
	if status == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Status parameter is required"})
		return
	}

	// Parse pagination parameters from the query
	limit, err := strconv.Atoi(c.DefaultQuery("limit", "10"))
	if err != nil || limit < 1 {
		limit = 10
	}

	page, err := strconv.Atoi(c.DefaultQuery("page", "1"))
	if err != nil || page < 1 {
		page = 1
	}
	offset := (page - 1) * limit

	// Fetch products by status with pagination
	products, err := controller.productService.GetProductsByStatusPaginated(status, limit, offset)
	if err != nil {
		log.Printf("GetProductsByStatus: failed to fetch products by status '%s': %v", status, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve products"})
		return
	}

	// Populate additional data and convert to ProductResponse
	var productResponses []models.ProductResponse
	for _, product := range products {
		productResponse, err := controller.populateAdditionalProductData(&product)
		if err != nil {
			log.Printf("GetProductsByStatus: failed to fetch additional data for product %s: %v", product.ID.String(), err)
			continue // Skip to the next product if there's an error
		}
		productResponses = append(productResponses, productResponse)
	}

	// Respond with the paginated products
	c.JSON(http.StatusOK, gin.H{"products": productResponses})
}

func (controller *ProductController) populateAdditionalTransactionData(product *models.ProductResponse) (models.DetailedProductResponse, error) {
	var productRes models.DetailedProductResponse

	// Fetch transactions for the product
	transactions, err := controller.TransactionService.GetByProductID(product.ID)
	if err != nil {
		return productRes, err
	}

	// Iterate over each transaction to fetch demographic information for the user
	var detailedTransactions []models.DetailedTransaction
	for _, transaction := range transactions {
		// Fetch the demographic information for the user involved in the transaction
		user, err := controller.UserService.GetDemographicInformation(transaction.UserID.String())
		if err != nil {
			return productRes, err
		}

		// Construct a detailed transaction with user demographic information
		detailedTransaction := models.DetailedTransaction{
			ID:          transaction.ID,
			ItemID:      transaction.ItemID,
			Description: transaction.Description,
			Action:      transaction.Action,
			ImageURL:    transaction.ImageURL,
			User:        *user, // Attach the user's demographic info
		}

		// Add to the slice of detailed transactions
		detailedTransactions = append(detailedTransactions, detailedTransaction)
	}

	productRes = models.DetailedProductResponse{
		User:          product.User.ID,
		ID:            product.ID,
		Name:          product.Name,
		Description:   product.Description,
		Price:         product.Price,
		Category:      product.Category,
		SubCategory:   product.SubCategory,
		RatingCount:   product.RatingCount,
		RatingAverage: product.RatingAverage,
		Rating:        product.Rating,
		CreatedAt:     product.CreatedAt,
		Status:        product.Status,
		Transactions:  detailedTransactions,
	}

	return productRes, nil
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

// GetRatedProductsByUserID godoc
// @Summary Get rated products by user ID
// @Description Fetches a list of products rated by the specified user
// @Tags Products
// @Accept json
// @Produce json
// @Param user_id query string true "User ID"
// @Success 200 {array} models.ProductResponse "List of rated products"
// @Router /products/rated [get]
func (controller *ProductController) GetRatedProductsByUserID(c *gin.Context) {
	userID := c.Query("user_id")

	// Fetch the ratings made by the user
	ratedItems, err := controller.RatingService.GetRatedProductIDsByUserID(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to fetch rated items"})
		return
	}

	// Iterate through the rated items and fetch product details for each
	var ratedProducts []models.ProductResponse
	for _, id := range ratedItems {
		productID, _ := uuid.Parse(id)
		product, err := controller.productService.GetByID(productID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to fetch product details"})
			return
		}

		// Append the product to the result
		p, _ := controller.populateAdditionalProductData(product)
		ratedProducts = append(ratedProducts, p)
	}

	// Return the list of rated products
	c.JSON(http.StatusOK, ratedProducts)
}

// GetPaginatedRandomProducts retrieves random products with pagination
// @Summary Get paginated random products
// @Tags         Products
// @Description Retrieve random products for unauthenticated users with pagination support
// @Param        count  query   int  true   "Number of products per page"
// @Param        page   query   int  true   "Page number"
// @Success 200 {array} models.ProductResponse
// @Router /products/random/paginated [get]
func (controller *ProductController) GetPaginatedRandomProducts(c *gin.Context) {
	// Get 'count' and 'page' query parameters from the request, with defaults if not specified
	count, err := strconv.Atoi(c.DefaultQuery("count", "10")) // Default to 10 products per page
	if err != nil || count <= 0 {
		count = 10
	}

	page, err := strconv.Atoi(c.DefaultQuery("page", "1")) // Default to the first page
	if err != nil || page <= 0 {
		page = 1
	}

	// Calculate the offset for pagination
	offset := (page - 1) * count

	// Fetch random paginated products from the product service
	products, err := controller.productService.GetRandomProductsPaginated(count, offset)
	if err != nil {
		log.Printf("GetPaginatedRandomProducts: failed to fetch random products: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve random products"})
		return
	}

	// Populate additional product data
	var productResponses []models.ProductResponse
	for _, product := range products {
		productResponse, err := controller.populateAdditionalProductData(&product)
		if err != nil {
			log.Printf("GetPaginatedRandomProducts: failed to fetch additional data for product %s: %v", product.ID.String(), err)
			continue // Skip to the next product if there's an error
		}
		productResponses = append(productResponses, productResponse)
	}

	// Send the paginated products in the response
	c.JSON(http.StatusOK, gin.H{
		"products": productResponses,
		"page":     page,
		"count":    count,
		"total":    len(products), // Total products in the current page
	})
}
