package db

import (
	"log"
	"sync"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"context-aware-ai/models"
)

var (
	db   *gorm.DB
	once sync.Once
)

func InitDB() {
	once.Do(func() {
		var err error
		db, err = gorm.Open(sqlite.Open("memory.db"), &gorm.Config{})
		if err != nil {
			log.Fatalf("Error connecting to the database: %v", err)
		}

		err = db.AutoMigrate(&models.Memory{}) 
		if err != nil {
			log.Fatalf("Error migrating the database: %v", err)
		}
	})
}

func GetDB() *gorm.DB {
	if db == nil {
		log.Fatal("Database not initialized. Call InitDB first.")
	}
	return db
}
