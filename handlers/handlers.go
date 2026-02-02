package handlers

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
	"context-aware-ai/db"
	"context-aware-ai/models" 
)

func storeMemory(memory models.Memory) error {
	return db.GetDB().Create(&memory).Error
}

func retrieveMemoryByLabel(label string) ([]models.Memory, error) {
	var memories []models.Memory
	err := db.GetDB().Where("label = ?", label).Find(&memories).Error
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
	return "This is a simulated response based on the query.", nil
}

func StoreMemoryHandler(c *gin.Context) {
	var memory models.Memory
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

func GetMemoryHandler(c *gin.Context) {
	label := c.Param("label")
	memories, err := retrieveMemoryByLabel(label)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve memory"})
		return
	}

	c.JSON(http.StatusOK, memories)
}

func QueryLLMHandler(c *gin.Context) {
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

	assistantResponse, err := CallLLM(prompt)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to query LLM"})
		return
	}

	err = storeMemory(models.Memory{
		Type:  "assistant_response",
		Label: "response_" + request.UserQuery,
		Value: assistantResponse,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to store assistant's response"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"response": assistantResponse})
}
