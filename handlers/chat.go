package handlers

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"
	"context-aware-ai/models"
	"context-aware-ai/services"
)

type ChatHandler struct {
	MemoryService *services.MemoryService
	OllamaService *services.OllamaService
	TopK          int
}

func (ch *ChatHandler) RunLoop() {
	reader := bufio.NewReader(os.Stdin)
	fmt.Println("Ollama Memory Chatbot (type 'quit' to exit)")
	for {
		fmt.Print("\nYou: ")
		userInput, _ := reader.ReadString('\n')
		userInput = strings.TrimSpace(userInput)
		if userInput == "" {
			continue
		}
		if strings.ToLower(userInput) == "quit" {
			fmt.Println("Goodbye!")
			break
		}
		queryEmbedding, err := ch.OllamaService.GetEmbedding(userInput)
		if err != nil {
			log.Println("Error embedding:", err)
			continue
		}
		memories, err := ch.MemoryService.RetrieveRelevant(queryEmbedding, ch.TopK)
		if err != nil {
			log.Println("Error retrieving memories:", err)
		}
		prompt := buildPrompt(userInput, memories)
		response, err := ch.OllamaService.GenerateResponse(prompt)
		if err != nil {
			log.Println("Error generating response:", err)
			continue
		}
		fmt.Println("\nAgent:", response)
		err = ch.MemoryService.StoreMemory(fmt.Sprintf("Q: %s A: %s", userInput, response), queryEmbedding)
		if err != nil {
			log.Println("Error storing memory:", err)
		}
	}
}

func buildPrompt(userInput string, memories []models.Memory) string {
	var sb strings.Builder
	sb.WriteString("Relevant Context:\n")
	for _, m := range memories {
		sb.WriteString("- ")
		sb.WriteString(m.Text)
		sb.WriteString("\n")
	}
	sb.WriteString("\nUser Question:\n")
	sb.WriteString(userInput)
	return sb.String()
}
