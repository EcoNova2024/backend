// models/user.go
package models

import (
	"github.com/google/uuid"
)

type User struct {
	ID           uuid.UUID `json:"id"`
	Username     string    `json:"username"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"password_hash"`
	ProfileImage string    `json:"profile_image"`
	Bio          string    `json:"bio"`
}
type SignUp struct {
	Username     string `json:"username" `
	Email        string `json:"email" `
	Password     string `json:"password" `
	ProfileImage string `json:"profile_image" `
	Bio          string `json:"bio" `
}

type Login struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}
