// backend/controller/user_controller.go
package controller

import (
	"backend/models"
	"backend/service"
	"net/http"
	"strconv"
	"time"

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
// @Description  Register a new user with provided user data.
// @Tags         Users
// @Accept       json
// @Produce      json
// @Param        user  body  models.SignUp  true  "User data for registration"
// @Success      201   {object} models.User              "User created successfully"
// @Failure      400   {object} map[string]string        "Invalid input"
// @Failure      500   {object} map[string]string        "Failed to create user"
// @Router       /users/signup [post]
func (controller *UserController) SignUp(c *gin.Context) {
	var user models.SignUp
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
// @Description  Authenticate a user and return a JWT token.
// @Tags         Users
// @Accept       json
// @Produce      json
// @Param        login  body  models.Login  true  "Login credentials for authentication"
// @Success      200    {object} map[string]interface{} "JWT token"
// @Failure      400    {object} map[string]string       "Invalid input"
// @Failure      401    {object} map[string]string       "Invalid credentials"
// @Router       /users/login [post]
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
	user, _ := controller.userService.GetByEmail(loginData.Email)

	c.JSON(http.StatusOK, gin.H{"token": token, "expires_at": time.Now().Add(3 * time.Hour), "user": user})
}

// GetDemographicInformation retrieves demographic information for a user
// @Summary      Get User Demographics
// @Description  Retrieve demographic information for a specific user by ID.
// @Tags         Users
// @Produce      json
// @Param        id  path  string  true  "User ID"
// @Success      200 {object} models.User              "User demographic information"
// @Failure      404 {object} map[string]string        "User not found"
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
// @Description  Update user information with provided user data.
// @Tags         Users
// @Accept       json
// @Produce      json
// @Param        user  body  models.UpdateUser  true  "User data for update"
// @Success      200   {object} models.User               "User updated successfully"
// @Failure      400   {object} map[string]string         "Invalid input"
// @Failure      401   {object} map[string]string         "User ID not found"
// @Failure      500   {object} map[string]string         "Failed to update user"
// @Router       /users [put]
func (controller *UserController) UpdateUser(c *gin.Context) {
	var user models.UpdateUser
	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input", "details": err.Error()})
		return
	}

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
// @Description  Update user's email address with a new email.
// @Tags         Users
// @Accept       json
// @Produce      json
// @Param        email  body  models.UpdateEmail  true  "New Email for update"
// @Success      200    {object} map[string]string       "Email updated successfully"
// @Failure      400    {object} map[string]string       "Invalid input"
// @Failure      401    {object} map[string]string       "User ID not found"
// @Failure      500    {object} map[string]string       "Failed to update email"
// @Router       /users/email [put]
func (controller *UserController) UpdateEmail(c *gin.Context) {
	var emailData models.UpdateEmail
	if err := c.ShouldBindJSON(&emailData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input", "details": err.Error()})
		return
	}

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

// UpdatePassword handles updating a user's password using a reset token
// @Summary      Update User Password
// @Description  Update user's password with a new password using a reset token.
// @Tags         Users
// @Accept       json
// @Produce      json
// @Param        token  query  string  true  "JWT token for user authentication"
// @Param        password  body  models.UpdatePassword  true  "New Password for update"
// @Success      200      {object} map[string]string      "Password updated successfully"
// @Failure      400      {object} map[string]string      "Invalid input"
// @Failure      401      {object} map[string]string      "Invalid or expired token"
// @Failure      500      {object} map[string]string      "Failed to update password"
// @Router       /users/password [put]
func (controller *UserController) UpdatePassword(c *gin.Context) {
	token := c.Query("token")
	var passwordData models.UpdatePassword

	// Validate the JWT token using the service layer
	userID, err := controller.userService.ValidateToken(token, "password_reset")
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	// Bind the new password
	if err := c.ShouldBindJSON(&passwordData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input", "details": err.Error()})
		return
	}

	// Update the user's password
	if err := controller.userService.UpdatePassword(userID, passwordData.NewPassword); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to reset password", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Password updated successfully"})
}

// SendPasswordResetEmail handles sending a password reset email
// @Summary      Send Password Reset Email
// @Description  Sends a password reset email to the user with provided email.
// @Tags         Users
// @Accept       json
// @Produce      json
// @Param        email  body  models.SendPasswordResetEmail  true  "User Email for password reset"
// @Success      200    {object} map[string]string          "Password reset email sent successfully"
// @Failure      400    {object} map[string]string          "Invalid input"
// @Failure      500    {object} map[string]string          "Failed to send reset email"
// @Router       /users/password/reset [post]
func (controller *UserController) SendPasswordResetEmail(c *gin.Context) {
	var emailData models.SendPasswordResetEmail
	if err := c.ShouldBindJSON(&emailData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input", "details": err.Error()})
		return
	}

	// Send the password reset email
	if err := controller.userService.SendPasswordResetEmail(emailData.Email); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to send reset email", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Password reset email sent successfully"})
}

// VerifyEmail handles verifying a user's email address
// @Summary      Verify User Email
// @Description  Verify the user's email address using a verification token.
// @Tags         Users
// @Accept       json
// @Produce      json
// @Param        token  query  string  true  "Verification token"
// @Success      200    {object} map[string]string       "Email verified successfully"
// @Failure      400    {object} map[string]string       "Invalid token"
// @Failure      500    {object} map[string]string       "Failed to verify email"
// @Router       /users/verify [post]
func (controller *UserController) VerifyEmail(c *gin.Context) {
	token := c.Query("token")

	// Verify the email using the token
	if err := controller.userService.VerifyEmail(token); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid token", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Email verified successfully"})
}

// SendEmailVerification handles sending an email verification link
// @Summary      Send Email Verification
// @Description  Sends an email verification link to the user's email
// @Tags         Users
// @Accept       json
// @Produce      json
// @Param        email  body  models.SendEmailVerification  true  "User Email"
// @Success      200    {object} map[string]string  "Verification email sent successfully"
// @Failure      400    {object} map[string]string  "Invalid input"
// @Failure      500    {object} map[string]string  "Failed to send verification email"
// @Router       /users/email/send-verification [post]
func (controller *UserController) SendEmailVerification(c *gin.Context) {
	var emailData models.SendEmailVerification
	if err := c.ShouldBindJSON(&emailData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input", "details": err.Error()})
		return
	}

	if err := controller.userService.SendEmailVerification(emailData.Email); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to send verification email", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Verification email sent successfully"})
}

// GetByName
// @Summary      Get users by name prefix
// @Description  Retrieves up to 10 users whose names start with the provided prefix.
// @Tags         Users
// @Accept       json
// @Produce      json
// @Param        name  query  string  true  "Name prefix to search for"
// @Success      200   {array}  models.User   "List of users"
// @Failure      400   {object}  map[string]string  "Bad Request"
// @Failure      500   {object}  map[string]string  "Internal Server Error"
// @Router       /users/search [get]
func (controller *UserController) GetByName(c *gin.Context) {
	name := c.Query("name") // Get the 'name' parameter from query string
	if name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Name parameter is required"})
		return
	}

	users, err := controller.userService.GetUsersByNamePrefix(name)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, users)
}

// GetUserByEmail godoc
// @Summary Get a user by email
// @Description Retrieves a user by their email address from query parameters
// @Tags Users
// @Accept  json
// @Produce  json
// @Param email query string true "Email Address"
// @Success 200 {object} models.User
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /users/email [get]
func (c *UserController) GetUserByEmail(ctx *gin.Context) {
	email := ctx.Query("email") // Get email from query parameter

	if email == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Email is required"})
		return
	}

	user, err := c.userService.GetByEmail(email)
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	ctx.JSON(http.StatusOK, user)
}

// AddPremiumDaysHandler adds premium days to a user's subscription
// @Summary Add premium days to a user's subscription
// @Description Extends or sets the premium subscription for a user by adding a given number of days
// @Tags Users
// @Accept json
// @Produce json
// @Param days query int true "Number of days to add"
// @Success 200 {object} map[string]interface{} "Success"
// @Failure 400 {object} map[string]interface{} "Invalid input"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /users/premium [put]
func (c *UserController) AddPremiumDaysHandler(ctx *gin.Context) {
	// Extract userID from context locals (middleware must set it)
	userID, exists := ctx.Get("user_id")
	if !exists {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "userID not found in request context"})
		return
	}

	// Parse the days from the query parameter
	daysParam := ctx.Query("days")
	days, err := strconv.Atoi(daysParam)
	if err != nil || days <= 0 {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid number of days"})
		return
	}

	// Call the service layer to add premium days
	updatedUser, err := c.userService.AddPremiumDays(userID.(string), days)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Return the updated user as a response
	ctx.JSON(http.StatusOK, gin.H{
		"message":      "Premium days added successfully",
		"user":         updatedUser,
		"premiumUntil": updatedUser.PremiumUntil,
	})
}
