// backend/repository/user_repository.go
package repository

import (
	"backend/models"
	"errors"

	"gorm.io/gorm"
)

type UserRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{db: db}
}

// Create inserts a new user into the database
func (repo *UserRepository) Create(user *models.User) error {
	return repo.db.Create(user).Error
}

// GetByEmail retrieves a user by their email
func (repo *UserRepository) GetByEmail(email string) (*models.User, error) {
	var user models.User
	if err := repo.db.Where("email = ?", email).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}

// GetByID retrieves a user by their ID
func (repo *UserRepository) GetByID(id string) (*models.User, error) {
	var user models.User
	if err := repo.db.First(&user, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}

// Update modifies an existing user's information
func (repo *UserRepository) Update(userID string, user *models.User) error {
	return repo.db.Model(&models.User{}).Where("id = ?", userID).Updates(user).Error
}

// UpdateEmail modifies an existing user's email
func (repo *UserRepository) UpdateEmail(userID, newEmail string) error {
	return repo.db.Model(&models.User{}).
		Where("id = ?", userID).
		Updates(map[string]interface{}{
			"email":    newEmail,
			"verified": false,
		}).Error
}

// UpdatePassword modifies an existing user's password
func (repo *UserRepository) UpdatePassword(userID, hashedPassword string) error {
	return repo.db.Model(&models.User{}).Where("id = ?", userID).Update("password", hashedPassword).Error
}

// VerifyEmail verifies a user's email address
func (repo *UserRepository) VerifyEmail(userID string) error {
	// Find the user by ID
	var user models.User
	if err := repo.db.First(&user, "id = ?", userID).Error; err != nil {
		return errors.New("user not found")
	}

	// Update the verified status
	user.Verified = true

	// Save changes to the database
	if err := repo.db.Save(&user).Error; err != nil {
		return errors.New("failed to update user")
	}

	return nil
}

// FindByNamePrefix finds users whose names start with the provided prefix (up to 10 users)
func (r *UserRepository) FindByNamePrefix(name string) ([]models.User, error) {
	var users []models.User
	result := r.db.Where("name LIKE ?", name+"%").Limit(10).Find(&users)
	if result.Error != nil {
		return nil, result.Error
	}
	return users, nil
}
