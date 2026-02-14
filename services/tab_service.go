package services

import (
	"context-aware-ai/models"
	"gorm.io/gorm"
)

type TabService struct {
	DB *gorm.DB
}

func NewTabService(db *gorm.DB) *TabService {
	db.AutoMigrate(&models.Tab{})
	return &TabService{DB: db}
}

func (s *TabService) CreateTab(userID uint, tabName string) (*models.Tab, error) {
	tab := models.Tab{
		UserID: userID,
		Name:   tabName,
	}
	err := s.DB.Create(&tab).Error
	return &tab, err
}

func (s *TabService) GetTabs(userID uint) ([]models.Tab, error) {
	var tabs []models.Tab
	err := s.DB.Where("user_id = ?", userID).Find(&tabs).Error
	return tabs, err
}

func (s *TabService) GetTabByID(tabID uint) (*models.Tab, error) {
	var tab models.Tab
	err := s.DB.First(&tab, tabID).Error
	return &tab, err
}
