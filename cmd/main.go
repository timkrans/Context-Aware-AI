package main

import (
	"fmt"
	"context-aware-ai/db"
	"context-aware-ai/handlers"
	"context-aware-ai/services"
	"runtime"
	"os"
)

func main() {
	if runtime.GOOS == "windows" {
		fmt.Println("Windows is not currently supported")
		os.Exit(1)
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
	}

	chatHandler.RunLoop()
}
