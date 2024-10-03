// backend/models/user.go
package models

import (
	"time"

	"github.com/google/uuid"
)

// User represents the user model
type User struct {
	ID        uuid.UUID `gorm:"type:uuid;default:uuid_generate_v4();primaryKey" json:"id"`
	FirstName string    `gorm:"not null" json:"first_name"`
	LastName  string    `gorm:"not null" json:"last_name"`
	Email     string    `gorm:"not null;unique" json:"email"`
	Password  string    `gorm:"not null" json:"password"`
	Verified  bool      `gorm:"not null" json:"verified"`
	CreatedAt time.Time `gorm:"default:current_timestamp" json:"created_at"`
	ImageURL  string    `gorm:"not null" json:"image_url"`
}

type SignUp struct {
	FirstName string `json:"first_name" binding:"required"`
	LastName  string `json:"last_name" binding:"required"`
	Email     string `json:"email" binding:"required,email"`
	Password  string `json:"password" binding:"required,min=8"`
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

// SendPasswordResetEmail represents the data for sending a password reset email
type SendPasswordResetEmail struct {
	Email string `json:"email" binding:"required,email"`
}

// VerifyEmail represents the data for verifying a user's email
type VerifyEmail struct {
	Token string `json:"token" binding:"required"`
}
