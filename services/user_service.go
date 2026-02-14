package services

import (
	"context-aware-ai/models"
	"gorm.io/gorm"
)

type UserService struct {
	DB *gorm.DB
}

func NewUserService(db *gorm.DB) *UserService {
	db.AutoMigrate(&models.User{})
	return &UserService{DB: db}
}

func (s *UserService) CreateUser(name string) (*models.User, error) {
	user := models.User{
		Name: name,
	}
	err := s.DB.Create(&user).Error
	return &user, err
}

func (s *UserService) GetUserByID(userID uint) (*models.User, error) {
	var user models.User
	err := s.DB.First(&user, userID).Error
	return &user, err
}

func (s *UserService) GetAllUsers() ([]models.User, error) {
	var users []models.User
	err := s.DB.Find(&users).Error
	return users, err
}
