package repository

import (
	"backend/models"
	"database/sql"
	"errors"

	"github.com/google/uuid"
)

// UserRepository defines the interface for user-related database operations
type UserRepository interface {
	FindUserByEmail(email string) (*models.User, error)
	CreateUser(signUp *models.SignUp) error
	FindUserByID(id uuid.UUID) (*models.User, error)
	UpdateUser(user *models.User) error
	DeleteUser(id uuid.UUID) error
}

// userRepository implements the UserRepository interface
type userRepository struct {
	db *sql.DB
}

// NewUserRepository creates a new instance of userRepository
func NewUserRepository(db *sql.DB) UserRepository {
	return &userRepository{db: db}
}

// CreateUser inserts a new user into the database
func (r *userRepository) CreateUser(signUp *models.SignUp) error {
	// Generate a new UUID for the user
	userID := uuid.New()

	query := `
        INSERT INTO users (id, username, email, password_hash, profile_image, bio) 
        VALUES (?, ?, ?, ?, ?, ?)`
	_, err := r.db.Exec(query, userID, signUp.Username, signUp.Email, signUp.Password, signUp.ProfileImage, signUp.Bio)
	if err != nil {
		return err
	}
	return nil
}

// FindUserByEmail retrieves a user by their email address
func (r *userRepository) FindUserByEmail(email string) (*models.User, error) {
	var user models.User
	query := `SELECT id, username, email, password_hash, profile_image, bio FROM users WHERE email = ?`
	err := r.db.QueryRow(query, email).Scan(&user.ID, &user.Username, &user.Email, &user.PasswordHash, &user.ProfileImage, &user.Bio)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil // No user found
		}
		return nil, err
	}
	return &user, nil
}

// FindUserByID retrieves a user by their ID
func (r *userRepository) FindUserByID(id uuid.UUID) (*models.User, error) {
	var user models.User
	query := `SELECT id, username, email, password_hash, profile_image, bio FROM users WHERE id = ?`
	err := r.db.QueryRow(query, id).Scan(&user.ID, &user.Username, &user.Email, &user.PasswordHash, &user.ProfileImage, &user.Bio)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil // No user found
		}
		return nil, err
	}
	return &user, nil
}

// UpdateUser updates an existing user's information
func (r *userRepository) UpdateUser(user *models.User) error {
	query := `
        UPDATE users 
        SET username = ?, email = ?, password_hash = ?, profile_image = ?, bio = ? 
        WHERE id = ?`
	_, err := r.db.Exec(query, user.Username, user.Email, user.PasswordHash, user.ProfileImage, user.Bio, user.ID)
	return err
}

// DeleteUser removes a user from the database
func (r *userRepository) DeleteUser(id uuid.UUID) error {
	query := `DELETE FROM users WHERE id = ?`
	_, err := r.db.Exec(query, id)
	return err
}
