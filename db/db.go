package db

import (
	"log"
	"context-aware-ai/models"
	"gorm.io/gorm/logger"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var DB *gorm.DB

//remove log will add it back when switch to vectorized db
func Init() {
	var err error
	DB, err = gorm.Open(sqlite.Open("memory.db"), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent),})
	if err != nil {
		log.Fatal("Failed to connect database:", err)
	}
	DB.AutoMigrate(
		&models.User{}, 
		&models.Tab{}, 
		&models.Memory{},
		&models.Document{},
	)
}
