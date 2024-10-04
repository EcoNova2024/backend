package service

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/dgrijalva/jwt-go"
	"golang.org/x/crypto/bcrypt"
)

// EmailConfig holds the configuration for sending emails
type EmailConfig struct {
	User     string
	Password string
	Host     string
	Port     string
}

// HashPassword hashes the password using bcrypt
func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err // Return error if hashing fails
	}
	return string(bytes), nil
}

// CheckPasswordHash checks if the provided password matches the hashed password
func CheckPasswordHash(password, hash string) bool {

	err := bcrypt.CompareHashAndPassword([]byte(password), []byte(password))
	return err == nil
}

// GenerateJWT generates a JWT token for the user
func GenerateJWT(userID string) (string, error) {
	jwtSecret := os.Getenv("JWT_SECRET") // Fetch secret from environment variable
	claims := jwt.MapClaims{
		"user_id": userID,
		"exp":     time.Now().Add(time.Hour * 72).Unix(), // Token valid for 72 hours
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, err := token.SignedString([]byte(jwtSecret)) // Use the secret from environment
	if err != nil {
		return "", err
	}
	return signedToken, nil
}

// ObfuscateEmail masks part of the email for privacy
func ObfuscateEmail(email string) string {
	parts := strings.Split(email, "@")
	if len(parts) != 2 {
		return email // Invalid email format
	}
	// Replace part of the email with asterisks
	obfuscated := parts[0][:1] + strings.Repeat("*", len(parts[0])-2) + parts[0][len(parts[0])-1:] + "@" + parts[1]
	return obfuscated
}

// GenerateEmailVerificationToken generates a JWT token for email verification
func GenerateEmailVerificationToken(userID string) (string, error) {
	jwtSecret := os.Getenv("JWT_SECRET") // Fetch secret from environment variable
	claims := jwt.MapClaims{
		"user_id": userID,
		"exp":     time.Now().Add(time.Hour * 24).Unix(), // Token valid for 24 hours
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, err := token.SignedString([]byte(jwtSecret)) // Use the secret from environment
	if err != nil {
		return "", err
	}
	return signedToken, nil
}

// LoadEmailConfig loads email configuration from environment variables
func LoadEmailConfig() EmailConfig {
	return EmailConfig{
		User:     os.Getenv("EMAIL_USER"),     // Sender's email address
		Password: os.Getenv("EMAIL_PASSWORD"), // Sender's email password
		Host:     os.Getenv("EMAIL_HOST"),     // SMTP server host
		Port:     os.Getenv("EMAIL_PORT"),     // SMTP server port
	}
}

// SendEmail sends an email using the specified parameters
func SendEmail(to string, subject string, body string) error {
	/*
		config := LoadEmailConfig()

		// Create the message
		message := fmt.Sprintf("Subject: %s\r\n\r\n%s", subject, body)

		// Set up authentication information.
		auth := smtp.PlainAuth("", config.User, config.Password, config.Host)

		// Send the email
		err := smtp.SendMail(config.Host+":"+config.Port, auth, config.User, []string{to}, []byte(message))
		if err != nil {
			return fmt.Errorf("failed to send email: %w", err)
		}*/
	fmt.Print(body)
	return nil
}

// SendVerificationEmail sends a verification email to the user
func SendVerificationEmail(email, verificationLink string) error {
	subject := "Email Verification"
	body := fmt.Sprintf("Click the following link to verify your email: %s", verificationLink)
	return SendEmail(email, subject, body)
}

// SendPasswordResetEmail sends a password reset email to the user
func SendPasswordResetEmail(email, resetLink string) error {
	subject := "Password Reset"
	body := fmt.Sprintf("Click the following link to reset your password: %s", resetLink)
	return SendEmail(email, subject, body)
}
