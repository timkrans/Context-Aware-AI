package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"log"
	"net/http"
	"time"
)

type Memory struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	Type      string    `json:"type"`
	Label     string    `json:"label"`
	Value     string    `json:"value"`
	Timestamp time.Time `json:"timestamp" gorm:"default:current_timestamp"`
}

var db *gorm.DB

func initDB() {
	var err error
	db, err = gorm.Open(sqlite.Open("memory.db"), &gorm.Config{})
	if err != nil {
		log.Fatalf("Error connecting to the database: %v", err)
	}

	err = db.AutoMigrate(&Memory{})
	if err != nil {
		log.Fatalf("Error migrating the database: %v", err)
	}
}

func storeMemory(memory Memory) error {
	return db.Create(&memory).Error
}

func retrieveMemoryByLabel(label string) ([]Memory, error) {
	var memories []Memory
	err := db.Where("label = ?", label).Find(&memories).Error
	return memories, err
}

func buildLLMPrompt(userQuery string) (string, error) {
	relevantMemories, err := retrieveMemoryByLabel(userQuery)
	if err != nil {
		return "", err
	}

	prompt := "Here are some relevant facts from memory:\n"
	for _, mem := range relevantMemories {
		prompt += fmt.Sprintf("- %s: %s\n", mem.Label, mem.Value)
	}
	prompt += "\nNow, based on these facts, please answer the user's query:\n"
	prompt += fmt.Sprintf("User: %s\nAssistant:", userQuery)

	return prompt, nil
}

func CallLLM(prompt string) (string, error) {
	//to do add the call to llm
	return "",nil
}

func storeMemoryHandler(c *gin.Context) {
	var memory Memory
	if err := c.ShouldBindJSON(&memory); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := storeMemory(memory); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to store memory"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Memory stored successfully"})
}

func getMemoryHandler(c *gin.Context) {
	label := c.Param("label")
	memories, err := retrieveMemoryByLabel(label)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve memory"})
		return
	}

	c.JSON(http.StatusOK, memories)
}

func queryLLMHandler(c *gin.Context) {
	var request struct {
		UserQuery string `json:"user_query"`
	}
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	prompt, err := buildLLMPrompt(request.UserQuery)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to build prompt"})
		return
	}

	assistantResponse, err := callOllama(prompt)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to query LLM"})
		return
	}

	storeMemory(Memory{
		Type:  "assistant_response",
		Label: "response_" + request.UserQuery,
		Value: assistantResponse,
	})

	c.JSON(http.StatusOK, gin.H{"response": assistantResponse})
}

func main() {
	initDB()

	r := gin.Default()

	r.POST("/memory", storeMemoryHandler)
	r.GET("/memory/:label", getMemoryHandler)
	r.POST("/query", queryLLMHandler)

	if err := r.Run(":8080"); err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
}
