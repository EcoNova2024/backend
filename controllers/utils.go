package controllers

import (
	"errors"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// GenerateJWTToken generates a JWT token for the user
func GenerateJWTToken(userID string) (string, error) {
	// Get the JWT secret from the environment
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		return "", errors.New("JWT secret is not set")
	}

	// Create the claims (payload) for the token
	claims := jwt.MapClaims{
		"user_id": userID,                           // Store user ID in the token
		"exp":     time.Now().Add(time.Hour).Unix(), // Token expiration time (1 hour)
	}

	// Create a new JWT token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Sign the token with the secret key
	return token.SignedString([]byte(jwtSecret))
}
