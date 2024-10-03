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
}

// NewProductController creates a new ProductController instance
func NewProductController(productService *service.ProductService, transactionService *service.TransactionService) *ProductController {
	return &ProductController{productService: productService, TransactionService: transactionService}
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
// @Param        product  body      models.Product  true  "Product data"
// @Success      201      {object}  models.ProductResponse
// @Router       /products [post]
func (controller *ProductController) Create(c *gin.Context) {
	var product models.ProductRequest
	if err := c.ShouldBindJSON(&product); err != nil {
		log.Printf("Create product: invalid input: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}

	if err := controller.productService.Create(&product); err != nil {
		log.Printf("Create product: service error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create product"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Product created successfully", "product": product})
}

// Update handles the updating of an existing product
// @Summary      Update an existing product
// @Description  Update an existing product with the provided data
// @Tags         Products
// @Accept       json
// @Produce      json
// @Param        product  body      models.Product  true  "Updated product data"
// @Success      200      {object}  models.ProductResponse
// @Router       /products [put]
func (controller *ProductController) Update(c *gin.Context) {
	var product models.Product
	if err := c.ShouldBindJSON(&product); err != nil {
		log.Printf("Update product: invalid input: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}

	if err := controller.productService.Update(&product); err != nil {
		log.Printf("Update product: service error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update product"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Product updated successfully", "product": product})
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
// @Summary      Get a product by ID
// @Description  Retrieve a product by its UUID
// @Tags         Products
// @Param        id   path      string  true  "Product UUID"
// @Success      200   {object}  models.ProductResponse @Router       /products/{id} [get]
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

	c.JSON(http.StatusOK, gin.H{"product": product})
}

// GetContentBased retrieves products based on content-based filtering
// @Summary      Get content-based products
// @Description  Retrieve products based on content, e.g., based on image URL
// @Tags         Products
// @Accept       json
// @Produce      json
// @Param        image_url query string true "Image URL for content-based filtering"
// @Success      200      {array}   models.Product
// @Router       /products/content-based [get]
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
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve content-based products"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"products": products})
}

// GetCollaborative retrieves products using a collaborative filtering approach with pagination
// @Summary      Get collaborative-based products
// @Description  Retrieve products based on collaborative filtering with pagination
// @Tags         Products
// @Accept       json
// @Produce      json
// @Param        page query int false "Page number"
// @Success      200   {array}   models.Product
// @Router       /products/collaborative [get]
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

	c.JSON(http.StatusOK, gin.H{"products": products})
}

// GetProductsByUserID retrieves products by user ID
// @Summary      Get products by user ID
// @Description  Retrieve products that belong to a specific user
// @Tags         Products
// @Produce      json
// @Success      200   {array}   models.Product
// @Router       /products/user [get]
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

	c.JSON(http.StatusOK, gin.H{"products": products})
}
