package services

import (
	"golang.org/x/crypto/bcrypt"
	"context-aware-ai/models"
	"gorm.io/gorm"
	"fmt"
)

type UserService struct {
	DB *gorm.DB
}

func NewUserService(db *gorm.DB) *UserService {
	db.AutoMigrate(&models.User{})
	return &UserService{DB: db}
}

func (s *UserService) CreateUser(name, password string) (*models.User, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %v", err)
	}

	user := models.User{
		UserName:    name,
		PasswordHash: string(hashedPassword),
	}

	err = s.DB.Create(&user).Error
	if err != nil {
		return nil, fmt.Errorf("failed to create user: %v", err)
	}

	return &user, nil
}

func (s *UserService) GetUserByUserName(username string) (*models.User, error) {
	var user models.User
	err := s.DB.Where("user_name = ?", username).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (s *UserService) GetUserByID(userID uint) (*models.User, error) {
	var user models.User
	err := s.DB.First(&user, userID).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (s *UserService) CheckPassword(userID uint, password string) (bool, error) {
	var user models.User
	err := s.DB.First(&user, userID).Error
	if err != nil {
		return false, err
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password))
	if err != nil {
		return false, nil
	}
	return true, nil
}
