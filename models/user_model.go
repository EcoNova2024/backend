// backend/models/user.go
package models

import (
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/google/uuid"
)

// User represents the user model
type User struct {
	ID           uuid.UUID `gorm:"type:uuid;default:uuid_generate_v4();primaryKey" json:"id"`
	Name         string    `gorm:"not null" json:"name"`
	Email        string    `gorm:"not null;unique" json:"email"`
	Password     string    `gorm:"not null" json:"password"`
	Verified     bool      `gorm:"not null" json:"verified"`
	CreatedAt    time.Time `gorm:"default:current_timestamp" json:"created_at"`
	ImageURL     string    `gorm:"not null" json:"image_url"`
	PremiumUntil string    `json:"premium_until"`
}

type SignUp struct {
	Name     string `json:"name" binding:"required"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`
	ImageURL string `gorm:"not null" json:"image_url"`
}

// Login represents the data required for user authentication
type Login struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

// UpdateEmail represents the data for updating a user's email
type UpdateEmail struct {
	NewEmail string `json:"new_email" binding:"required,email"`
}

// UpdatePassword represents the data for updating a user's password
type UpdatePassword struct {
	NewPassword string `json:"new_password" binding:"required,min=6"`
}

type UpdateUser struct {
	NewUser  string `json:"new_user"`
	NewImage string `json:"new_image"`
}

// SendPasswordResetEmail represents the data for sending a password reset email
type SendPasswordResetEmail struct {
	Email string `json:"email" binding:"required,email"`
}

// VerifyEmail represents the data for verifying a user's email
type VerifyEmail struct {
	Token string `json:"token" binding:"required"`
}
type SendEmailVerification struct {
	Email string `json:"email" binding:"required,email"`
}
type PasswordResetClaims struct {
	UserID string `json:"user_id"`
	jwt.StandardClaims
}
