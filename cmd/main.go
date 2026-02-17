package main

import (
	"context-aware-ai/db"
	"context-aware-ai/handlers"
	"context-aware-ai/services"
	"log"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"os"
	"github.com/gin-contrib/cors"
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
	var llmService services.LLMService
	llmProvider := os.Getenv("LLM_PROVIDER")
	switch llmProvider {
	case "openai":
		return &OpenAIService{
			APIKey: os.Getenv("OPENAI_API_KEY"),
			Model:  os.Getenv("OPENAI_MODEL"),
		}
	case "claude":
		return &ClaudeService{
			APIKey: os.Getenv("CLAUDE_API_KEY"),
			Model:  os.Getenv("CLAUDE_MODEL"),
		}
	case "gemini":
		return &GeminiService{
			APIKey: os.Getenv("GEMINI_API_KEY"),
			Model:  os.Getenv("GEMINI_MODEL"),
		}
	case "ollama":
		llmService = ollamaService
	default:
		log.Fatal("Unknown LLM provider")
	}

	chatHandler := &handlers.ChatHandler{
		MemoryService: memoryService,
		TabService:    tabService,
		UserService:   userService,
		OllamaService: ollamaService,
		LLMService:    llmService,
		TopK:          3,
		JWTSecret:     []byte(jwtSecretKey),
	}

	r := gin.Default()
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"}, // Allow all origins (you can specify specific domains if needed)
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
	}))

	chatHandler.SetupRoutes(r)
	if err := r.Run(":3000"); err != nil {
		log.Fatal(err)
	}
}
