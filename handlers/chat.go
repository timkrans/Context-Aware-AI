package handlers

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"
	"context-aware-ai/models"
	"context-aware-ai/services"
	"golang.org/x/crypto/ssh/terminal"
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

	var username string
	fmt.Print("Enter your username (or type 'new' to create a new user): ")
	usernameInput, _ := reader.ReadString('\n')
	usernameInput = strings.TrimSpace(usernameInput)

	var user *models.User

	if usernameInput == "new" {
		fmt.Print("Enter new username: ")
		username, _ = reader.ReadString('\n')
		username = strings.TrimSpace(username)

		fmt.Print("Enter password: ")
		//the terminal method only works on linux and mac so this a problem for window 
		//TODO fix for windows
		passwordBytes, err := terminal.ReadPassword(int(os.Stdin.Fd()))
		if err != nil {
			log.Println("Error reading password:", err)
			return
		}
		password := string(passwordBytes)
		fmt.Println()
		user, err = ch.UserService.CreateUser(username, password)
		if err != nil {
			log.Println("Error creating user:", err)
			return
		}
		if user == nil {
			log.Println("User creation failed: user is nil")
			return
		}
		fmt.Println("Created new user:", user.UserName)
	} else {
		username = usernameInput
		//have to declare error to make it so you dont use :=
		var err error
		user, err = ch.UserService.GetUserByUserName(username)
		if err != nil {
			log.Println("Error retrieving user:", err)
			return
		}
		fmt.Println("Selected user:", user.UserName)
		fmt.Print("Enter password: ")
		//the terminal method only works on linux and mac so this a problem for window 
		//TODO fix for windows
		passwordBytes, err := terminal.ReadPassword(int(os.Stdin.Fd()))
		if err != nil {
			log.Println("Error reading password:", err)
			return
		}
		password := string(passwordBytes)
		fmt.Println()

		valid, err := ch.UserService.CheckPassword(user.ID, password)
		if err != nil {
			log.Println("Error checking password:", err)
			return
		}
		if !valid {
			log.Println("Invalid password")
			return
		}
		fmt.Println("Password verified successfully.")
	}

	if user == nil || user.ID == 0 {
		log.Println("Error: User ID is invalid")
		return
	}

	tabs, err := ch.TabService.GetTabs(user.ID)
	if err != nil {
		log.Println("Error getting tabs:", err)
		return
	}
	if tabs == nil {
		log.Println("No tabs found for user.")
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
		selectedTab, err = ch.TabService.CreateTab(user.ID, tabName)
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

	ch.startChatLoop(reader, user.ID, selectedTab.ID)
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

		if strings.ToLower(userInput) == "logout" {
			fmt.Println("logging out...")
			ch.RunLoop()
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
