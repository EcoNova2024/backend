package controllers

import (
	"backend/models"
	"log"
	"strings"

	"github.com/gofiber/fiber/v2"
	"golang.org/x/crypto/bcrypt"
)

// UserRepository defines the interface for user repository
type UserRepository interface {
	FindUserByEmail(email string) (*models.User, error)
	CreateUser(signUp *models.SignUp) error
}

// UserController handles user-related operations
type UserController struct {
	repo UserRepository
}

// NewUserController initializes a new UserController with a UserRepository
func NewUserController(repo UserRepository) *UserController {
	return &UserController{repo: repo}
}

// SignUp handles user registration
// @Summary      User Registration
// @Description  Register a new user
// @Tags         Users
// @Accept       json
// @Produce      json
// @Param        user  body  models.SignUp  true  "User data"
// @Success      201   {object} models.User
// @Router       /signup [post]
func (c *UserController) SignUp(ctx *fiber.Ctx) error {
	var signUpData models.SignUp

	// Parse the request body
	if err := ctx.BodyParser(&signUpData); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid input"})
	}

	// Check if the user already exists
	if existingUser, _ := c.repo.FindUserByEmail(signUpData.Email); existingUser != nil {
		return ctx.Status(fiber.StatusConflict).JSON(fiber.Map{"error": "Email already in use"})
	}

	// Hash the password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(signUpData.Password), bcrypt.DefaultCost)
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Could not hash password"})
	}
	signUpData.Password = string(hashedPassword)

	// Create the user
	if err := c.repo.CreateUser(&signUpData); err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Could not create user"})
	}

	return ctx.Status(fiber.StatusCreated).JSON(signUpData)
}

// Login handles user authentication
// @Summary      User Login
// @Description  Authenticate a user and return user information
// @Tags         Users
// @Accept       json
// @Produce      json
// @Param        login  body  models.Login  true  "Login credentials"
// @Success      200    {object} models.User
// @Router       /login [post]
// Login handles user authentication
// Login handles user authentication
func (c *UserController) Login(ctx *fiber.Ctx) error {
	var login models.Login

	// Parse the request body
	if err := ctx.BodyParser(&login); err != nil {
		log.Printf("Error parsing body: %v", err)
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid input"})
	}

	// Trim the input
	login.Email = strings.TrimSpace(login.Email)
	login.Password = strings.TrimSpace(login.Password)

	// Find the user by email
	user, err := c.repo.FindUserByEmail(login.Email)
	if err != nil {
		log.Printf("Error during user lookup: %v", err)
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Internal server error"})
	}

	// Check if user exists
	if user == nil {
		return ctx.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Invalid credentials"})
	}

	// Compare the provided password with the stored hash

	// bcrypt.CompareHashAndPassword compares the stored hashed password with the provided plaintext password
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(login.Password)); err != nil {
		log.Printf("Invalid password for user: %s", user.Username)
		return ctx.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Invalid credentials"})
	}

	return ctx.JSON(user) // Return user information
}
