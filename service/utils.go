package service

import (
	"fmt"
	"net/smtp"
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
	// Replace $2y$ with $2a$ for compatibility
	if strings.HasPrefix(hash, "$2y$") {
		hash = "$2a$" + hash[4:] // Replace $2y$ with $2a$
	}

	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

// GenerateJWT generates a JWT token for the user with purpose and expiration
func GenerateJWT(userID, purpose string, expiresIn time.Duration) (string, error) {
	jwtSecret := os.Getenv("JWT_SECRET") // Fetch secret from environment variable
	claims := jwt.MapClaims{
		"user_id": userID,
		"exp":     time.Now().Add(expiresIn).Unix(), // Token valid for specified duration
		"purpose": purpose,
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
	return GenerateJWT(userID, "email_verification", time.Hour*24) // Token valid for 24 hours
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

// SendEmail sends an HTML email
func SendEmail(to, subject, htmlBody string) error {
	config := LoadEmailConfig()

	// Create MIME headers for HTML email
	headers := make(map[string]string)
	headers["From"] = config.User
	headers["To"] = to
	headers["Subject"] = subject
	headers["MIME-Version"] = "1.0"
	headers["Content-Type"] = "text/html; charset=\"UTF-8\""

	// Format headers and body
	message := ""
	for k, v := range headers {
		message += fmt.Sprintf("%s: %s\r\n", k, v)
	}
	message += "\r\n" + htmlBody

	// Set up authentication information
	auth := smtp.PlainAuth("", config.User, config.Password, config.Host)

	// Send the email
	err := smtp.SendMail(config.Host+":"+config.Port, auth, config.User, []string{to}, []byte(message))
	if err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}

	fmt.Println("Email sent successfully")
	return nil
}

func SendVerifyEmail(email, verificationLink string) error {
	subject := "Verify Your Email Address"
	htmlBody := fmt.Sprintf(`
        <html>
        <body style="font-family: Arial, sans-serif; background-color: #f4f4f4; padding: 20px;">
            <div style="max-width: 600px; margin: auto; background-color: #ffffff; padding: 20px; border-radius: 10px; box-shadow: 0 4px 8px rgba(0, 0, 0, 0.1);">
                <h2 style="text-align: center; color: #2c3e50;">Welcome to Our Community!</h2>
                <p style="color: #555; line-height: 1.6;">
                    Thank you for signing up. Please verify your email address by clicking the button below:
                </p>
                <div style="text-align: center; margin: 30px 0;">
                    <a href="%s" style="display: inline-block; padding: 12px 24px; background-color: #4CAF50; color: #ffffff; text-decoration: none; border-radius: 5px; font-weight: bold;">Verify Email</a>
                </div>
                <p style="color: #555; line-height: 1.6;">
                    If you did not create this account, you can safely ignore this email.
                </p>
                <hr style="border: none; border-top: 1px solid #ddd; margin: 20px 0;">
                <p style="text-align: center; color: #aaa; font-size: 12px;">&copy; 2024 Renova, Inc. All rights reserved.</p>
            </div>
        </body>
        </html>`, verificationLink)

	return SendEmail(email, subject, htmlBody)
}

func SendResetEmail(email, resetLink string) error {
	subject := "Reset Your Password"
	htmlBody := fmt.Sprintf(`
        <html>
        <body style="font-family: Arial, sans-serif; background-color: #f4f4f4; padding: 20px;">
            <div style="max-width: 600px; margin: auto; background-color: #ffffff; padding: 20px; border-radius: 10px; box-shadow: 0 4px 8px rgba(0, 0, 0, 0.1);">
                <h2 style="text-align: center; color: #e74c3c;">Password Reset Request</h2>
                <p style="color: #555; line-height: 1.6;">
                    We received a request to reset your password. Click the button below to reset it:
                </p>
                <div style="text-align: center; margin: 30px 0;">
                    <a href="%s" style="display: inline-block; padding: 12px 24px; background-color: #e74c3c; color: #ffffff; text-decoration: none; border-radius: 5px; font-weight: bold;">Reset Password</a>
                </div>
                <p style="color: #555; line-height: 1.6;">
                    If you did not request a password reset, you can ignore this email.
                </p>
                <hr style="border: none; border-top: 1px solid #ddd; margin: 20px 0;">
                <p style="text-align: center; color: #aaa; font-size: 12px;">&copy; 2024 Renova, Inc. All rights reserved.</p>
            </div>
        </body>
        </html>`, resetLink)

	return SendEmail(email, subject, htmlBody)
}

// GeneratePasswordResetToken generates a JWT token for password reset
func GeneratePasswordResetToken(userID string) (string, error) {
	return GenerateJWT(userID, "password_reset", time.Hour) // Token valid for 1 hour
}
