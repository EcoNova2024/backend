// backend/service/user_service.go
package service

import (
	"backend/models"
	"backend/repository"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/dgrijalva/jwt-go"
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

func (service *UserService) Create(req *models.SignUp) error {
	// Create a new user with the provided information
	user := &models.User{
		Email:     req.Email,
		Name:      req.Name,
		CreatedAt: time.Now().UTC(),
		Verified:  false,
		ImageURL:  req.ImageURL,
	}

	// Hash the password and handle any errors
	hashedPassword, err := HashPassword(req.Password)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}
	user.Password = hashedPassword

	// Store the new user in the repository
	if err := service.userRepo.Create(user); err != nil {
		return fmt.Errorf("failed to create user: %w", err)
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
	user, err := service.userRepo.GetByID(id)
	if err != nil {
		return nil, ErrUserNotFound
	}

	// Obfuscate sensitive information
	user.Password = ""
	user.Email = (user.Email)
	return user, nil
}

func (service *UserService) UpdateUser(userID string, req *models.UpdateUser) error {
	// Retrieve existing user data
	user, err := service.userRepo.GetByID(userID)
	if err != nil {
		return ErrUserNotFound
	}

	// Update user information
	user.Name = req.NewUser
	user.ImageURL = req.NewImage

	// Persist updated user data
	return service.userRepo.Update(userID, user)
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
	resetLink := fmt.Sprintf("https://yourfrontend.com/reset-password?token=%s", resetToken)

	// Prepare email content
	subject := "Password Reset Request"
	body := fmt.Sprintf("To reset your password, click the following link: %s", resetLink)

	// Send the email
	if err = SendEmail(user.Email, subject, body); err != nil {
		return errors.New("failed to send password reset email")
	}

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

	// Prepare email content
	subject := "Email Verification"
	body := fmt.Sprintf("Please verify your email by clicking the following link: %s", verificationLink)

	// Send the email
	if err = SendEmail(user.Email, subject, body); err != nil {
		return errors.New("failed to send verification email")
	}

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

	return users, nil
}

func (s *UserService) GetByEmail(email string) (*models.User, error) {
	user, err := s.userRepo.GetByEmail(email)
	if err != nil {
		return nil, err
	}

	// Set the password to an empty string
	user.Password = ""

	return user, nil
}
