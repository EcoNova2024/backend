package routes

import (
	"backend/controller"
	"backend/middleware" // Import JWT middleware
	"backend/repository"
	"backend/service"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// SetupRoutes initializes the routes for the Gin router
func SetupRoutes(router *gin.Engine, db *gorm.DB) {
	// Create repositories using the repository factory
	repoFactory := repository.NewRepositoryFactory(db)
	productRepo := repoFactory.GetProductRepository()
	ratingRepo := repoFactory.GetRatingRepository()
	userRepo := repoFactory.GetUserRepository()
	transactionRepo := repoFactory.GetTransactionRepository()
	commentRepo := repoFactory.GetCommentRepository() // Add comment repository

	// Create services
	productService := service.NewProductService(productRepo)
	ratingService := service.NewRatingService(ratingRepo)
	userService := service.NewUserService(userRepo)
	transactionService := service.NewTransactionService(transactionRepo)
	commentService := service.NewCommentService(commentRepo) // Create comment service

	// Create controllers
	productController := controller.NewProductController(productService, transactionService, userService, ratingService)
	ratingController := controller.NewRatingController(ratingService)
	userController := controller.NewUserController(userService)
	homeController := controller.NewHomeController()
	transactionController := controller.NewTransactionController(transactionService, productService)
	commentController := controller.NewCommentController(commentService, *userService)

	// Define routes
	router.GET("/", homeController.Index) // Home route

	// User routes
	users := router.Group("/users")
	{
		users.POST("/signup", userController.SignUp)                                 // DONE!
		users.POST("/login", userController.Login)                                   // DONE!
		users.GET("/:id", userController.GetDemographicInformation)                  // DONE!
		users.PUT("/", middleware.JWTAuth(), userController.UpdateUser)              // DONE!
		users.PUT("/email", middleware.JWTAuth(), userController.UpdateEmail)        // DONE!
		users.PUT("/password", userController.UpdatePassword)                        // DONE!
		users.POST("/password/reset", userController.SendPasswordResetEmail)         // DONE!
		users.POST("/verify", userController.VerifyEmail)                            // DONE!
		users.POST("/email/send-verification", userController.SendEmailVerification) // DONE!
		users.GET("/search", userController.GetByName)
		users.GET("/email", userController.GetUserByEmail)
		users.PUT("/premium", middleware.JWTAuth(), userController.AddPremiumDaysHandler)
	}

	// Product routes
	products := router.Group("/products")
	{
		products.POST("/", middleware.JWTAuth(), productController.Create)                       // Create a new product
		products.GET("/", productController.GetOne)                                              // Get a product by ID
		products.GET("/user", productController.GetProductsByUserID)                             // Get products by user ID (from JWT)
		products.GET("/content-based", productController.GetContentBased)                        // Get content-based recommendations
		products.GET("/collaborative", middleware.JWTAuth(), productController.GetCollaborative) // Get collaborative-based recommendations
		products.GET("/status", productController.GetProductsByStatus)                           // Get restored products
		products.GET("/random", productController.GetRandomProducts)                             // Get random products
		products.GET("/rated", productController.GetRatedProductsByUserID)
		products.GET("/random/paginated", productController.GetPaginatedRandomProducts)
		products.GET("/item-based", productController.GetItemBased)
	}

	// Rating routes
	ratings := router.Group("/ratings")
	{
		ratings.POST("/", middleware.JWTAuth(), ratingController.Create)                          // Create a new rating
		ratings.DELETE("/:id", middleware.JWTAuth(), ratingController.Delete)                     // Delete a rating by ID
		ratings.GET("/user/:user_id", ratingController.GetRatedProductsByUserId)                  // Get all rated products by user ID
		ratings.GET("/product/:product_id/average", ratingController.GetAverageRatingByProductId) // Get average rating and count by product ID
	}

	// Transaction routes
	transactions := router.Group("/transactions")
	{
		transactions.POST("/:item_id/", middleware.JWTAuth(), transactionController.AddTransactionToItem) // Add transaction to item
	}

	// Comment routes
	comments := router.Group("/comments")
	{
		comments.POST("/", middleware.JWTAuth(), commentController.Create)      // Create comment
		comments.GET("/product/:product_id", commentController.GetByProductID)  // Get comments by product
		comments.DELETE("/:id", middleware.JWTAuth(), commentController.Delete) // Delete comment
	}
}
