// backend/routes/setup_routes.go
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
	productController := controller.NewProductController(productService, transactionService, userService, ratingService)
	ratingController := controller.NewRatingController(ratingService)
	userController := controller.NewUserController(userService)
	commentsController := controller.NewCommentsController(commentsService)
	homeController := controller.NewHomeController()
	transactionController := controller.NewTransactionController(transactionService, productService)

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
	}

	// Product routes
	products := router.Group("/products")
	{
		products.POST("/", middleware.JWTAuth(), productController.Create) // Create a new product
		products.GET("/:id", productController.GetOne)                     // Get a product by ID
		products.GET("/user", productController.GetProductsByUserID)       // Get products by user ID (from JWT)
		products.GET("/content-based", productController.GetContentBased)  // Get content-based recommendations
		products.GET("/collaborative", productController.GetCollaborative) // Get collaborative-based recommendations
	}

	// Rating routes
	ratings := router.Group("/ratings")
	{
		ratings.POST("/:user_id/:product_id", ratingController.Create)                            // Create a new rating
		ratings.DELETE("/:id", ratingController.Delete)                                           // Delete a rating by ID
		ratings.GET("/user/:user_id", ratingController.GetRatedProductsByUserId)                  // Get all rated products by user ID
		ratings.GET("/product/:product_id/average", ratingController.GetAverageRatingByProductId) // Get average rating and count by product ID
	}

	// Comments routes
	comments := router.Group("/comments")
	{
		comments.POST("", commentsController.Create)                                 // Create a new comment
		comments.POST("/:comment_id/hide", commentsController.HideComment)           // Hide a comment
		comments.GET("/user/:user_id", commentsController.GetCommentsByUserId)       // Get comments by user ID
		comments.GET("/product/:product_id", commentsController.GetCommentsByItemId) // Get comments by product ID
	}

	// Transaction routes
	transactions := router.Group("/transactions")
	{
		transactions.POST("/:item_id/", middleware.JWTAuth(), transactionController.AddTransactionToItem) // Add transaction to item
	}
}
