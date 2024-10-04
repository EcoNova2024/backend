// backend/service/user_service.go
package service

import (
	"backend/models"
	"backend/repository"
	"errors"
	"fmt"
	"time"
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

func (service *UserService) UpdatePassword(userID, newPassword string) error {
	hashedPassword, err := HashPassword(newPassword)
	if err != nil {
		return ErrInternal
	}
	return service.userRepo.UpdatePassword(userID, hashedPassword)
}

func (service *UserService) SendPasswordResetEmail(email string) error {
	return nil
}

func (service *UserService) VerifyEmail(token string) error {
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
	verificationLink := fmt.Sprintf("https://yourfrontend.com/verify-email?token=%s", verificationToken)

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
