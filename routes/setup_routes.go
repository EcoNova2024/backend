// backend/routes/setup_routes.go
package routes

import (
	"backend/controller"
	"backend/repository" // Make sure to import the repository package
	"backend/service"    // Import the service package

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// SetupRoutes initializes the routes for the Gin router
func SetupRoutes(router *gin.Engine, db *gorm.DB) {
	// Create repositories
	repoFactory := repository.NewRepositoryFactory(db)
	productRepo := repoFactory.GetProductRepository()
	ratingRepo := repoFactory.GetRatingRepository()
	userRepo := repoFactory.GetUserRepository()
	commentsRepo := repoFactory.GetCommentsRepository()
	transactionRepo := repoFactory.GetTransactionRepository()

	// Create services
	productService := service.NewProductService(productRepo)
	ratingService := service.NewRatingService(ratingRepo)
	userService := service.NewUserService(userRepo)
	commentsService := service.NewCommentsService(commentsRepo)
	transactionService := service.NewTransactionService(transactionRepo)

	// Create controllers
	productController := controller.NewProductController(productService, transactionService)
	ratingController := controller.NewRatingController(ratingService)
	userController := controller.NewUserController(userService)
	commentsController := controller.NewCommentsController(commentsService)
	homeController := controller.NewHomeController()
	transactionController := controller.NewTransactionController(transactionService)

	// Define routes
	router.GET("/", homeController.Index)              // Home route
	router.POST("/products", productController.Create) // Create a new product
	router.POST("/ratings", ratingController.Create)   // Create a new rating
	router.POST("/users", userController.SignUp)       // Create a new user
	comments := router.Group("/comments")
	{
		comments.POST("", commentsController.Create)                                 // Create a new comment
		comments.POST("/:comment_id/hide", commentsController.HideComment)           // Hide a comment
		comments.GET("/user/:user_id", commentsController.GetCommentsByUserId)       // Get comments by user ID
		comments.GET("/product/:product_id", commentsController.GetCommentsByItemId) // Get comments by product ID
	}
	products := router.Group("/products")
	{
		// Create a new product
		products.POST("/", productController.Create)

		// Update an existing product
		products.PUT("/", productController.Update)

		// Delete a product by ID
		products.DELETE("/:id", productController.Delete)

		// Get a product by ID
		products.GET("/:id", productController.GetOne)

		// Get products by user ID (from JWT)
		products.GET("/user", productController.GetProductsByUserID)

		// Get products based on content-based recommendations
		products.GET("/content-based", productController.GetContentBased)

		// Get collaborative-based recommendations with pagination
		products.GET("/collaborative", productController.GetCollaborative)
	}

	ratings := router.Group("/ratings")
	{
		ratings.POST("/:user_id/:product_id", ratingController.Create)                            // Create a new rating
		ratings.DELETE("/:id", ratingController.Delete)                                           // Delete a rating by ID
		ratings.GET("/user/:user_id", ratingController.GetRatedProductsByUserId)                  // Get all rated products by user ID
		ratings.GET("/product/:product_id/average", ratingController.GetAverageRatingByProductId) // Get average rating and count by product ID
	}
	users := router.Group("/users")
	{
		users.POST("/signup", userController.SignUp)                         // User Registration
		users.POST("/login", userController.Login)                           // User Login
		users.GET("/:id", userController.GetDemographicInformation)          // Get User Demographics
		users.PUT("/", userController.UpdateUser)                            // Update User Information
		users.PUT("/email", userController.UpdateEmail)                      // Update User Email
		users.PUT("/password", userController.UpdatePassword)                // Update User Password
		users.POST("/password/reset", userController.SendPasswordResetEmail) // Send Password Reset Email
		users.POST("/email/verify", userController.VerifyEmail)              // Verify User Email
	}
	transactions := router.Group("/transactions")
	{
		transactions.POST("/item/:item_id/user/:user_id", transactionController.AddTransactionToItem) // Add transaction to item
		transactions.PATCH("/:id/hide", transactionController.HideTransaction)                        // Hide transaction
	}
}
