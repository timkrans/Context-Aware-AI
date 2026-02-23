package main

import (
	"context-aware-ai/db"
	"context-aware-ai/handlers"
	"context-aware-ai/services"
	"log"
	"github.com/gin-gonic/gin"
	"context-aware-ai/loadenv"
	"os"
)

func main() {
	//added my own loading to help keep depencies to a minimum
	_ = loadenv.LoadEnv("")
	//took out the other check because it breaks if dockerized 
	//but this will check if set
	jwtSecretKey := os.Getenv("JWT_SECRET_KEY")
	if jwtSecretKey == "" {
		log.Fatal("JWT_SECRET_KEY not set in .env file")
	}

	db.Init()
	memoryService := &services.MemoryService{DB: db.DB}
	tabService := &services.TabService{DB: db.DB}
	userService := &services.UserService{DB: db.DB}
	ollamaService := &services.OllamaService{
		BaseURL:        os.Getenv("OLLAMA_BASE_URL"),
		GenerateModel:  os.Getenv("OLLAMA_GENERATE_MODEL"),
		EmbeddingModel: os.Getenv("OLLAMA_EMBEDDING_MODEL"),
	}
	ragService := &services.RAGService{ DB: db.DB, OllamaService: ollamaService, }
	var llmService services.LLMService
	llmProvider := os.Getenv("LLM_PROVIDER")
	switch llmProvider {
	case "openai":
		llmService = &services.OpenAIService{
			APIKey: os.Getenv("OPENAI_API_KEY"),
			Model:  os.Getenv("OPENAI_MODEL"),
		}
		break
	case "claude":
		llmService = &services.ClaudeService{
			APIKey: os.Getenv("CLAUDE_API_KEY"),
			Model:  os.Getenv("CLAUDE_MODEL"),
		}
		break
	case "gemini":
		llmService = &services.GeminiService{
			APIKey: os.Getenv("GEMINI_API_KEY"),
			Model:  os.Getenv("GEMINI_MODEL"),
		}
		break
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
		RAGService :	ragService,
		TopK:          3,
		JWTSecret:     []byte(jwtSecretKey),
	}

	r := gin.Default()

	//for when the frontend is created
	r.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	})

	chatHandler.SetupRoutes(r)
	fileHandler := &handlers.FileHandler{
		RAGService:  ragService,
		ChatHandler: chatHandler,
	}
	fileHandler.SetupRoutes(r)
	if err := r.Run(":3000"); err != nil {
		log.Fatal(err)
	}
}
