package main

import (
	"github.com/gin-gonic/gin"
	"log"
	"context-aware-ai/db"
	"context-aware-ai/handlers"
)

func main() {
	db.InitDB()

	r := gin.Default()

	r.POST("/memory", handlers.StoreMemoryHandler)
	r.GET("/memory/:label", handlers.GetMemoryHandler)
	r.POST("/query", handlers.QueryLLMHandler)

	if err := r.Run(":8080"); err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
}
