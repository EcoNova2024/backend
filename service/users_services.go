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
	user := &models.User{
		Email:     req.Email,
		Name:      req.Name,
		CreatedAt: time.Now().UTC(),
		Verified:  false,
		ImageURL:  req.ImageURL,
	}

	hashedPassword, err := HashPassword(req.Password)
	if err != nil {
		return ErrInternal
	}
	user.Password = hashedPassword

	if err := service.userRepo.Create(user); err != nil {
		return ErrInternal
	}

	return nil
}

func (service *UserService) Authenticate(email, password string) (string, error) {
	user, err := service.userRepo.GetByEmail(email)
	if err != nil {
		return "", ErrUserNotFound
	}
	if user == nil {
		return "", ErrUserNotFound
	}

	if !CheckPasswordHash(password, user.Password) {
		return "", ErrInvalidCredentials
	}

	token, err := GenerateJWT(user.ID.String())
	if err != nil {
		return "", ErrInternal
	}

	return token, nil
}

func (service *UserService) GetDemographicInformation(id string) (*models.User, error) {
	user, err := service.userRepo.GetByID(id)
	if err != nil {
		return nil, ErrUserNotFound
	}
	user.Password = ""
	user.Email = ObfuscateEmail(user.Email)
	return user, nil
}

func (service *UserService) UpdateUser(userID string, req *models.UpdateUser) error {
	user, err := service.userRepo.GetByID(userID)
	if err != nil {
		return ErrUserNotFound
	}
	user.Name = req.NewUser
	user.ImageURL = req.NewImage
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
	err = SendEmail(user.Email, subject, body)
	if err != nil {
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

	// Generate verification token (could be a JWT or a random string)
	verificationToken, err := GenerateEmailVerificationToken(user.ID.String())
	if err != nil {
		return errors.New("failed to generate verification token")
	}

	// Create verification link
	verificationLink := fmt.Sprintf("https://%s/verify-email?token=%s", os.Getenv("FE_PORT"), verificationToken)

	// Prepare email content
	subject := "Email Verification"
	body := fmt.Sprintf("Please verify your email by clicking the following link: %s", verificationLink)

	// Send the email (use a utility function or an external service to send the email)
	err = SendEmail(user.Email, subject, body)
	if err != nil {
		return errors.New("failed to send verification email")
	}

	return nil
}
func (service *UserService) UpdatePassword(userID, newPassword string) error {
	hashedPassword, err := HashPassword(newPassword)
	if err != nil {
		return ErrInternal
	}
	return service.userRepo.UpdatePassword(userID, hashedPassword)
}

// ValidateToken checks if the reset token is a valid JWT and extracts the user ID
func (service *UserService) ValidateToken(token string) (string, error) {
	jwtSecret := os.Getenv("JWT_SECRET") // Fetch secret from environment variable

	// Parse the token
	parsedToken, err := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("invalid signing method")
		}
		return []byte(jwtSecret), nil
	})

	if err != nil || !parsedToken.Valid {
		return "", errors.New("invalid token")
	}

	// Extract user ID from claims
	if claims, ok := parsedToken.Claims.(jwt.MapClaims); ok {
		userID := claims["user_id"].(string) // Extract user_id from JWT claims
		return userID, nil
	}

	return "", errors.New("invalid token claims")
}
func (service *UserService) VerifyEmail(token string) error {
	// Validate the token and extract user ID
	userID, err := service.ValidateToken(token)
	if err != nil {
		return errors.New("invalid or expired token")
	}

	// Update user's verification status
	if err := service.userRepo.VerifyEmail(userID); err != nil {
		return errors.New("failed to verify email")
	}

	return nil
}
