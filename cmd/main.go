package main

import (
	"context-aware-ai/db"
	"context-aware-ai/handlers"
	"context-aware-ai/services"
	"log"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"os"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Fatal("Error loading .env file")
	}

	jwtSecretKey := os.Getenv("JWT_SECRET_KEY")
	if jwtSecretKey == "" {
		log.Fatal("JWT_SECRET_KEY not set in .env file")
	}

	db.Init()

	memoryService := &services.MemoryService{DB: db.DB}
	tabService := &services.TabService{DB: db.DB}
	userService := &services.UserService{DB: db.DB}
	ollamaService := &services.OllamaService{
		BaseURL:        "http://localhost:11434",
		GenerateModel:  "llama3.2",
		EmbeddingModel: "nomic-embed-text",
	}

	chatHandler := &handlers.ChatHandler{
		MemoryService: memoryService,
		TabService:    tabService,
		UserService:   userService,
		OllamaService: ollamaService,
		TopK:          3,
		JWTSecret:     []byte(jwtSecretKey),
	}

	r := gin.Default()
	chatHandler.SetupRoutes(r)
	if err := r.Run(":3000"); err != nil {
		log.Fatal(err)
	}
}
