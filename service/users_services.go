package service

import (
	"backend/models"
	"backend/repository"
	"encoding/base64"
	"errors"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/google/uuid"
)

var (
	// Authentication errors
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrEmailNotVerified   = errors.New("email not verified")
	ErrProductNotFound    = errors.New("product not found")
	ErrUserNotFound       = errors.New("user not found")
	// Validation errors
	ErrInvalidInput       = errors.New("invalid input")
	ErrEmailAlreadyExists = errors.New("email already exists")

	// Token errors
	ErrInvalidToken = errors.New("invalid token")
	ErrTokenExpired = errors.New("token has expired")

	// Internal errors
	ErrInternal = errors.New("internal server error")
)

type UserService struct {
	userRepo *repository.UserRepository
}

func NewUserService(userRepo *repository.UserRepository) *UserService {
	return &UserService{userRepo: userRepo}
}

// Handle image settings (pre-signed URL generation and image URL updates)
func (service *UserService) handleImage(user *models.User) error {
	// Check if an image URL exists and generate a pre-signed URL if needed
	if user.ImageURL != "" {
		// Construct the S3 object key for the user's image
		imageKey := fmt.Sprintf("users/%s", user.ImageURL)

		// Use the GetImage utility to get the pre-signed URL
		_, err := GetImage(imageKey)
		if err != nil {
			return fmt.Errorf("failed to retrieve image URL: %v", err)
		}

		// Replace the ImageURL with the pre-signed URL
		user.ImageURL = imageKey
	} else {
		// If no image URL is provided, set it to an empty string
		user.ImageURL = ""
	}
	return nil
}

func (service *UserService) Create(req *models.SignUp) error {
	// Create a new user with the provided information
	user := &models.User{
		ID:        uuid.New(),
		Email:     req.Email,
		Name:      req.Name,
		CreatedAt: time.Now().UTC(),
		Verified:  false,
		ImageURL:  req.ImageURL, // Initial value from request
	}

	// Hash the password and handle any errors
	hashedPassword, err := HashPassword(req.Password)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}
	user.Password = hashedPassword

	// Handle image settings (generate pre-signed URL if image exists)
	if req.ImageURL != "" {
		// Log the image handling process
		log.Printf("Handling image for user: %s", user.ID.String())
		// Decode and upload image if base64 data is provided
		imageData, err := base64.StdEncoding.DecodeString(req.ImageURL)
		if err != nil {
			log.Printf("Error decoding base64 image data for user %s: %v", user.Email, err)
			return fmt.Errorf("failed to decode image data: %v", err)
		}

		// Generate a unique key for the image based on the user ID (or another identifier)
		imageKey := fmt.Sprintf("user-images/%s.jpg", user.Email)

		// Upload the image to S3 and get the pre-signed URL
		_, err = PutImage(imageKey, imageData)
		if err != nil {
			log.Printf("Error uploading image for user %s: %v", user.Email, err)
			return fmt.Errorf("failed to upload image: %v", err)
		}

		// Set the image URL in the user object
		user.ImageURL = imageKey
		log.Printf("Successfully uploaded image for user %s, URL: %s", user.Email, imageKey)
	}

	// Store the new user in the repository
	if err := service.userRepo.Create(user); err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}

	return nil
}

func (service *UserService) UpdateUser(userID string, req *models.UpdateUser) error {
	// Retrieve existing user data
	user, err := service.userRepo.GetByID(userID)
	if err != nil {
		return ErrUserNotFound
	}

	// Update user information, only if new values are provided
	if req.NewUser != "" {
		user.Name = req.NewUser // Only update if new name is provided
	}

	if req.NewImage != "" {
		user.ImageURL = req.NewImage // Only update if new image URL is provided
	}

	// Handle image settings (generate pre-signed URL if image exists)
	if req.NewImage != "" {
		// Log the image handling process
		log.Printf("Handling image for user ID: %s", userID)

		// Decode and upload image if base64 data is provided
		imageData, err := base64.StdEncoding.DecodeString(req.NewImage)
		if err != nil {
			log.Printf("Error decoding base64 image data for user ID %s: %v", userID, err)
			return fmt.Errorf("failed to decode image data: %v", err)
		}

		// Generate a unique key for the image based on the user ID
		imageKey := fmt.Sprintf("%s.jpg", userID)

		// Upload the image to S3 and get the pre-signed URL
		imageURL, err := PutImage("users/"+imageKey, imageData)
		if err != nil {
			log.Printf("Error uploading image for user ID %s: %v", userID, err)
			return fmt.Errorf("failed to upload image: %v", err)
		}

		// Set the image URL in the user object
		user.ImageURL = imageKey
		log.Printf("Successfully uploaded image for user ID %s, URL: %s", userID, imageURL)
	}

	// Persist updated user data
	if err := service.userRepo.Update(userID, user); err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}

	return nil
}

func (service *UserService) Authenticate(email, password string) (string, error) {
	// Retrieve user by email
	user, err := service.userRepo.GetByEmail(email)
	if err != nil || user == nil {
		return "", ErrInvalidCredentials
	}

	// Validate the password
	if !CheckPasswordHash(password, user.Password) {
		return "", ErrInvalidCredentials
	}

	// Generate JWT token for authentication
	token, err := GenerateJWT(user.ID.String(), "auth", 3*time.Hour)
	if err != nil {
		return "", fmt.Errorf("failed to generate JWT: %w", err)
	}

	return token, nil
}

func (service *UserService) GetDemographicInformation(id string) (*models.User, error) {
	// Fetch user by ID from the repository
	user, err := service.userRepo.GetByID(id)
	if err != nil {
		return nil, ErrUserNotFound
	}

	// Obfuscate sensitive information
	user.Password = ""                      // Clear password
	user.Email = ObfuscateEmail(user.Email) // Optional email obfuscation

	// Handle image settings (generate pre-signed URL if image exists)
	if err := service.handleImage(user); err != nil {
		return nil, fmt.Errorf("failed to handle image settings: %v", err)
	}

	// Return the user object with the pre-signed URL (if available)
	return user, nil
}

func (service *UserService) UpdateEmail(userID, newEmail string) error {
	return service.userRepo.UpdateEmail(userID, newEmail)
}

func (service *UserService) SendPasswordResetEmail(email string) error {
	// Check if the user exists
	user, err := service.userRepo.GetByEmail(email)
	if err != nil {
		return errors.New("user not found")
	}

	// Generate a password reset token
	resetToken, err := GeneratePasswordResetToken(user.ID.String())
	if err != nil {
		return errors.New("failed to generate password reset token")
	}

	// Create the reset link
	resetLink := fmt.Sprintf("https://%s/verify-email?token=%s", os.Getenv("FE_PORT"), resetToken)

	SendResetEmail(email, resetLink)

	return nil
}

func (service *UserService) SendEmailVerification(email string) error {
	// Check if the user exists
	user, err := service.userRepo.GetByEmail(email)
	if err != nil {
		return errors.New("user not found")
	}

	// Generate verification token
	verificationToken, err := GenerateEmailVerificationToken(user.ID.String())
	if err != nil {
		return errors.New("failed to generate verification token")
	}

	// Create verification link
	verificationLink := fmt.Sprintf("https://%s/verify-email?token=%s", os.Getenv("FE_PORT"), verificationToken)

	SendVerifyEmail(email, verificationLink)

	return nil
}

func (service *UserService) UpdatePassword(userID, newPassword string) error {
	hashedPassword, err := HashPassword(newPassword)
	if err != nil {
		return fmt.Errorf("failed to hash new password: %w", err)
	}
	return service.userRepo.UpdatePassword(userID, hashedPassword)
}

// ValidateToken checks if the reset token is a valid JWT and extracts the user ID
func (service *UserService) ValidateToken(token string, expectedPurpose string) (string, error) {
	jwtSecret := os.Getenv("JWT_SECRET") // Fetch secret from environment variable

	// Parse the token
	parsedToken, err := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("invalid signing method")
		}
		return []byte(jwtSecret), nil
	})

	if err != nil {
		return "", fmt.Errorf("failed to parse token: %w", err)
	}

	if !parsedToken.Valid {
		return "", ErrInvalidToken
	}

	// Extract claims
	if claims, ok := parsedToken.Claims.(jwt.MapClaims); ok {
		// Check purpose
		if purpose, ok := claims["purpose"].(string); !ok || purpose != expectedPurpose {
			return "", errors.New("token purpose does not match expected purpose")
		}

		// Check if the token is expired
		if exp, ok := claims["exp"].(float64); ok {
			expirationTime := time.Unix(int64(exp), 0) // Convert expiration to time.Time
			if time.Now().After(expirationTime) {
				return "", ErrTokenExpired // Return error if token is expired
			}
		} else {
			return "", errors.New("expiration time not found in token claims")
		}

		// Extract user ID
		if userID, ok := claims["user_id"].(string); ok {
			return userID, nil
		}
		return "", errors.New("user ID not found in token claims")
	}

	return "", errors.New("invalid token claims")
}

func (service *UserService) VerifyEmail(token string) error {
	// Validate the token and extract user ID
	userID, err := service.ValidateToken(token, "email_verification")
	if err != nil {
		return errors.New("invalid or expired token")
	}

	// Update user's verification status
	if err := service.userRepo.VerifyEmail(userID); err != nil {
		return errors.New("failed to verify email")
	}

	return nil
}

// GetUsersByNamePrefix retrieves users whose names start with the given prefix (up to 10 users)
func (s *UserService) GetUsersByNamePrefix(name string) ([]models.User, error) {
	users, err := s.userRepo.FindByNamePrefix(name)
	if err != nil {
		return nil, err
	}

	// Set the password to an empty string for each user
	for i := range users {
		users[i].Password = ""
	}

	// Handle image settings (generate pre-signed URL if image exists)
	for i := range users {
		if err := s.handleImage(&users[i]); err != nil {
			return nil, fmt.Errorf("failed to handle image for user %s: %v", users[i].ID, err)
		}
	}

	return users, nil
}

func (s *UserService) GetByEmail(email string) (*models.User, error) {
	user, err := s.userRepo.GetByEmail(email)
	if err != nil {
		return nil, err
	}

	// Set the password to an empty string
	user.Password = ""

	// Handle image settings (generate pre-signed URL if image exists)
	if err := s.handleImage(user); err != nil {
		return nil, fmt.Errorf("failed to handle image for user %s: %v", user.ID, err)
	}

	return user, nil
}
func (s *UserService) AddPremiumDays(userID string, days int) (*models.User, error) {
	if days <= 0 {
		return nil, errors.New("days must be a positive integer")
	}

	// Call the repository method to update premium
	updatedUser, err := s.userRepo.AddPremiumByDay(userID, days)
	if err != nil {
		return nil, err
	}

	return updatedUser, nil
}
