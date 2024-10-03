// backend/service/user_service.go
package service

import (
	"backend/models"
	"backend/repository"
	"errors"
)

type UserService struct {
	userRepo *repository.UserRepository
}

func NewUserService(userRepo *repository.UserRepository) *UserService {
	return &UserService{userRepo: userRepo}
}

func (service *UserService) Create(user *models.User) error {
	hashedPassword, err := HashPassword(user.Password)
	if err != nil {
		return err
	}
	user.Password = hashedPassword
	return service.userRepo.Create(user)
}

func (service *UserService) Authenticate(email, password string) (string, error) {
	user, err := service.userRepo.GetByEmail(email)
	if err != nil {
		return "", err
	}

	if !CheckPasswordHash(password, user.Password) {
		return "", errors.New("invalid credentials")
	}

	token, err := GenerateJWT(user.ID.String())
	if err != nil {
		return "", err
	}

	return token, nil
}

func (service *UserService) GetDemographicInformation(id string) (*models.User, error) {
	user, err := service.userRepo.GetByID(id)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (service *UserService) UpdateUser(userID string, user *models.User) error {
	return service.userRepo.Update(userID, user)
}

func (service *UserService) UpdateEmail(userID, newEmail string) error {
	return service.userRepo.UpdateEmail(userID, newEmail)
}

func (service *UserService) UpdatePassword(userID, newPassword string) error {
	hashedPassword, err := HashPassword(newPassword)
	if err != nil {
		return err
	}
	return service.userRepo.UpdatePassword(userID, hashedPassword)
}

func (service *UserService) SendPasswordResetEmail(email string) error {
	// USE SENT MAIL UTIL
	return nil
}

func (service *UserService) VerifyEmail(token string) error {
	// VERIFY IT BY PARSEING JWT
	return nil
}
