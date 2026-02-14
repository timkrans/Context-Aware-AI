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
	TabService    *services.TabService
	UserService   *services.UserService
	OllamaService *services.OllamaService
	TopK          int
}

func (ch *ChatHandler) RunLoop() {
	reader := bufio.NewReader(os.Stdin)
	fmt.Println("Ollama Memory Chatbot (type 'quit' to exit)")

	var userID uint
	fmt.Print("Enter User ID (or type 'new' to create a new user): ")
	userInput, _ := reader.ReadString('\n')
	userInput = strings.TrimSpace(userInput)

	if userInput == "new" {
		fmt.Print("Enter new user name: ")
		name, _ := reader.ReadString('\n')
		name = strings.TrimSpace(name)
		user, err := ch.UserService.CreateUser(name)
		if err != nil {
			log.Println("Error creating user:", err)
			return
		}
		fmt.Println("Created new user:", user.Name)
		userID = user.ID
	} else {
		fmt.Sscanf(userInput, "%d", &userID)
		user, err := ch.UserService.GetUserByID(userID)
		if err != nil {
			log.Println("Error retrieving user:", err)
			return
		}
		fmt.Println("Selected user:", user.Name)
	}

	tabs, err := ch.TabService.GetTabs(userID)
	if err != nil {
		log.Println("Error getting tabs:", err)
		return
	}

	fmt.Println("Your Tabs:")
	for i, tab := range tabs {
		fmt.Printf("%d: %s\n", i+1, tab.Name)
	}

	fmt.Print("Select a tab by number or type 'new' to create a new tab: ")
	tabInput, _ := reader.ReadString('\n')
	tabInput = strings.TrimSpace(tabInput)

	var selectedTab *models.Tab
	if tabInput == "new" {
		fmt.Print("Enter new tab name: ")
		tabName, _ := reader.ReadString('\n')
		tabName = strings.TrimSpace(tabName)
		selectedTab, err = ch.TabService.CreateTab(userID, tabName)
		if err != nil {
			log.Println("Error creating new tab:", err)
			return
		}
		fmt.Println("Created new tab:", selectedTab.Name)
	} else {
		var tabIndex int
		fmt.Sscanf(tabInput, "%d", &tabIndex)
		if tabIndex < 1 || tabIndex > len(tabs) {
			fmt.Println("Invalid tab selection.")
			return
		}
		selectedTab = &tabs[tabIndex-1]
	}

	ch.startChatLoop(reader, userID, selectedTab.ID)
}

func (ch *ChatHandler) startChatLoop(reader *bufio.Reader, userID uint, tabID uint) {
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

		memories, err := ch.MemoryService.RetrieveRelevant(queryEmbedding, ch.TopK, userID, tabID)
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

		err = ch.MemoryService.StoreMemory(fmt.Sprintf("Q: %s A: %s", userInput, response), queryEmbedding, userID, tabID)
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
