// backend/controller/user_controller.go
package controller

import (
	"backend/models"
	"backend/service"
	"net/http"

	"github.com/gin-gonic/gin"
)

// UserController handles HTTP requests related to users
type UserController struct {
	userService *service.UserService
}

// NewUserController creates a new UserController instance
func NewUserController(userService *service.UserService) *UserController {
	return &UserController{userService: userService}
}

// SignUp handles user registration or creation
// @Summary      User Registration
// @Description  Register a new user
// @Tags         Users
// @Accept       json
// @Produce      json
// @Param        user  body  models.SignUp  true  "User data"
// @Success      201   {object} models.User
// @Router       /signup [post]
func (controller *UserController) SignUp(c *gin.Context) {
	var user models.User
	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input", "details": err.Error()})
		return
	}

	// Create the user
	if err := controller.userService.Create(&user); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user", "details": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "User created successfully", "user": user})
}

// Login handles user authentication
// @Summary      User Login
// @Description  Authenticate a user and return a JWT token
// @Tags         Users
// @Accept       json
// @Produce      json
// @Param        login  body  models.Login  true  "Login credentials"
// @Success      200    {object} map[string]interface{} "token"
// @Router       /login [post]
func (controller *UserController) Login(c *gin.Context) {
	var loginData models.Login
	if err := c.ShouldBindJSON(&loginData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input", "details": err.Error()})
		return
	}

	token, err := controller.userService.Authenticate(loginData.Email, loginData.Password)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"token": token})
}

// GetDemographicInformation retrieves demographic information for a user
// @Summary      Get User Demographics
// @Description  Retrieve demographic information for a specific user
// @Tags         Users
// @Produce      json
// @Param        id  path  string  true  "User ID"
// @Success      200 {object} models.User
// @Router       /users/{id} [get]
func (controller *UserController) GetDemographicInformation(c *gin.Context) {
	id := c.Param("id")

	user, err := controller.userService.GetDemographicInformation(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"user": user})
}

// UpdateUser handles updating user information
// @Summary      Update User
// @Description  Update user information
// @Tags         Users
// @Accept       json
// @Produce      json
// @Param        user  body  models.User  true  "User data"
// @Router       /users [put]
func (controller *UserController) UpdateUser(c *gin.Context) {
	var user models.User
	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input", "details": err.Error()})
		return
	}

	// Retrieve user ID from context locals
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User ID not found in context"})
		return
	}

	if err := controller.userService.UpdateUser(userID.(string), &user); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update user", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "User updated successfully", "user": user})
}

// UpdateEmail handles updating a user's email address
// @Summary      Update User Email
// @Description  Update user's email address
// @Tags         Users
// @Accept       json
// @Produce      json
// @Param        email  body  models.UpdateEmail  true  "New Email"
// @Router       /users/email [put]
func (controller *UserController) UpdateEmail(c *gin.Context) {
	var emailData models.UpdateEmail
	if err := c.ShouldBindJSON(&emailData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input", "details": err.Error()})
		return
	}

	// Retrieve user ID from context locals
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User ID not found in context"})
		return
	}

	if err := controller.userService.UpdateEmail(userID.(string), emailData.NewEmail); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update email", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Email updated successfully"})
}

// UpdatePassword handles updating a user's password
// @Summary      Update User Password
// @Description  Update user's password
// @Tags         Users
// @Accept       json
// @Produce      json
// @Param        password  body  models.UpdatePassword  true  "New Password"
// @Router       /users/password [put]
func (controller *UserController) UpdatePassword(c *gin.Context) {
	var passwordData models.UpdatePassword
	if err := c.ShouldBindJSON(&passwordData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input", "details": err.Error()})
		return
	}

	// Retrieve user ID from context locals
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User ID not found in context"})
		return
	}

	if err := controller.userService.UpdatePassword(userID.(string), passwordData.NewPassword); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update password", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Password updated successfully"})
}

// SendPasswordResetEmail handles sending a password reset email
// @Summary      Send Password Reset Email
// @Description  Sends a password reset email to the user
// @Tags         Users
// @Accept       json
// @Produce      json
// @Param        email  body  models.SendPasswordResetEmail  true  "User Email"
// @Router       /users/password/reset [post]
func (controller *UserController) SendPasswordResetEmail(c *gin.Context) {
	var emailData models.SendPasswordResetEmail
	if err := c.ShouldBindJSON(&emailData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input", "details": err.Error()})
		return
	}

	if err := controller.userService.SendPasswordResetEmail(emailData.Email); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to send reset email", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Password reset email sent successfully"})
}

// VerifyEmail handles email verification
// @Summary      Verify User Email
// @Description  Verify the user's email using a token
// @Tags         Users
// @Accept       json
// @Produce      json
// @Param        token  body  models.VerifyEmail  true  "Verification Token"
// @Router       /users/email/verify [post]
func (controller *UserController) VerifyEmail(c *gin.Context) {
	var verifyData models.VerifyEmail
	if err := c.ShouldBindJSON(&verifyData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input", "details": err.Error()})
		return
	}

	if err := controller.userService.VerifyEmail(verifyData.Token); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to verify email", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Email verified successfully"})
}
