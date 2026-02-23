package handlers

import (
	"fmt"
	"strings"
	"net/http"
	"context-aware-ai/models"
	"context-aware-ai/services"
	"github.com/gin-gonic/gin"
	"github.com/dgrijalva/jwt-go"
	"time"
	"strconv"
)

type ChatHandler struct {
	MemoryService *services.MemoryService
	TabService    *services.TabService
	LLMService    services.LLMService 
	UserService   *services.UserService
	OllamaService *services.OllamaService
	RAGService   *services.RAGService
	TopK          int
	JWTSecret     []byte
}

func (ch *ChatHandler) SetupRoutes(router *gin.Engine) {
	router.POST("/create-user", ch.CreateUserHandler)
	router.POST("/login", ch.LoginHandler)
	router.POST("/refresh-token", ch.RefreshTokenHandler)
	router.GET("/tabs", ch.GetTabsHandler)
	router.POST("/tabs", ch.CreateTabHandler)
	router.DELETE("/tabs/:id", ch.DeleteTabHandler)
	router.POST("/chat", ch.ChatHandler)
}

func (ch *ChatHandler) CreateUserHandler(c *gin.Context) {
	var input struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}

	user, err := ch.UserService.CreateUser(input.Username, input.Password)
	if err != nil || user == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error creating user"})
		return
	}

	c.JSON(http.StatusCreated, user)
}

func (ch *ChatHandler) GenerateSessionToken(user *models.User) (string, error) {
	expirationTime := time.Now().Add(24 * time.Hour)
	claims := &jwt.StandardClaims{
		Issuer:    fmt.Sprintf("%d", user.ID),
		ExpiresAt: expirationTime.Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	sessionToken, err := token.SignedString(ch.JWTSecret)
	if err != nil {
		return "", err
	}
	return sessionToken, nil
}

func (ch *ChatHandler) GenerateRefreshToken(user *models.User) (string, error) {
	expirationTime := time.Now().Add(30 * 24 * time.Hour)
	claims := &jwt.StandardClaims{
		Issuer:    fmt.Sprintf("%d", user.ID),
		ExpiresAt: expirationTime.Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	refreshToken, err := token.SignedString(ch.JWTSecret)
	if err != nil {
		return "", err
	}
	return refreshToken, nil
}

func (ch *ChatHandler) Authenticate(c *gin.Context) (*models.User, error) {
	sessionToken := c.GetHeader("Authorization")
	if sessionToken == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Missing session token"})
		return nil, fmt.Errorf("missing session token")
	}
	tokenString := strings.TrimPrefix(sessionToken, "Bearer ")
	claims := &jwt.StandardClaims{}
	_, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return ch.JWTSecret, nil
	})
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid session token"})
		return nil, fmt.Errorf("invalid session token")
	}

	userID := claims.Issuer
	userIDInt, err := strconv.ParseUint(userID, 10, 32)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid user ID"})
		return nil, fmt.Errorf("invalid user ID")
	}

	user, err := ch.UserService.GetUserByID(uint(userIDInt))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return nil, err
	}
	return user, nil
}

func (ch *ChatHandler) LoginHandler(c *gin.Context) {
	var input struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}

	user, err := ch.UserService.GetUserByUserName(input.Username)
	if err != nil || user == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	valid, err := ch.UserService.CheckPassword(user.ID, input.Password)
	if err != nil || !valid {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid password"})
		return
	}

	accessToken, err := ch.GenerateSessionToken(user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error generating access token"})
		return
	}

	refreshToken, err := ch.GenerateRefreshToken(user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error generating refresh token"})
		return
	}

	c.Header("Authorization", "Bearer "+accessToken)
	c.JSON(http.StatusOK, gin.H{
		"access_token":  accessToken,
		"refresh_token": refreshToken,
	})
}

func (ch *ChatHandler) RefreshTokenHandler(c *gin.Context) {
	var input struct {
		RefreshToken string `json:"refresh_token"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid refresh token"})
		return
	}

	claims := &jwt.StandardClaims{}
	refreshTokenString := input.RefreshToken
	_, err := jwt.ParseWithClaims(refreshTokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return ch.JWTSecret, nil
	})
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid refresh token"})
		return
	}

	userID := claims.Issuer
	userIDInt, err := strconv.ParseUint(userID, 10, 32)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid user ID"})
		return
	}

	user, err := ch.UserService.GetUserByID(uint(userIDInt))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	accessToken, err := ch.GenerateSessionToken(user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error generating new access token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"access_token": accessToken,
	})
}

func (ch *ChatHandler) GetTabsHandler(c *gin.Context) {
	user, err := ch.Authenticate(c)
	if err != nil {
		return
	}

	tabs, err := ch.TabService.GetTabs(user.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error getting tabs"})
		return
	}

	c.JSON(http.StatusOK, tabs)
}

func (ch *ChatHandler) CreateTabHandler(c *gin.Context) {
	user, err := ch.Authenticate(c)
	if err != nil {
		return
	}

	var input struct {
		TabName string `json:"tab_name"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}

	tab, err := ch.TabService.CreateTab(user.ID, input.TabName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error creating tab"})
		return
	}

	c.JSON(http.StatusCreated, tab)
}

func (ch *ChatHandler) DeleteTabHandler(c *gin.Context) {
    user, err := ch.Authenticate(c)
    if err != nil {
        return
    }

    tabIDStr := c.Param("id")
    tabID, err := strconv.Atoi(tabIDStr)
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid tab ID"})
        return
    }

    tabs, err := ch.TabService.GetTabs(user.ID)
    if err != nil || tabID < 1 || tabID > len(tabs) {
        c.JSON(http.StatusNotFound, gin.H{"error": "Tab not found"})
        return
    }

    tab := tabs[tabID-1]
    err = ch.MemoryService.DeleteMemoriesByTabID(user.ID, tab.ID)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Error deleting memories"})
        return
    }

	if err := ch.RAGService.DeleteDocumentsByTabID(user.ID, tab.ID); err != nil { 
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error deleting documents"})
		return 
	}

    err = ch.TabService.DeleteTab(user.ID, tab.ID)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Error deleting tab"})
        return
    }

    c.JSON(http.StatusOK, gin.H{"message": "Tab, memories, and documents deleted successfully"})
}

func (ch *ChatHandler) ChatHandler(c *gin.Context) {
    var input struct {
        TabID   uint   `json:"tab_id"`
        Message string `json:"message"`
		//optional to add reason to the chat
		Reasoning *bool `json:"reasoning"`
    }

    if err := c.ShouldBindJSON(&input); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
        return
    }

    user, err := ch.Authenticate(c)
    if err != nil {
        return
    }

    tabs, err := ch.TabService.GetTabs(user.ID)
    if err != nil || len(tabs) == 0 {
        c.JSON(http.StatusNotFound, gin.H{"error": "No tabs found"})
        return
    }

    if input.TabID < 1 || input.TabID > uint(len(tabs)) {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid TabID"})
        return
    }

    tab := tabs[input.TabID-1]
    input.TabID = tab.ID

    queryEmbedding, err := ch.OllamaService.GetEmbedding(input.Message)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Error embedding message"})
        return
    }

    memories, err := ch.MemoryService.RetrieveRelevant(queryEmbedding, ch.TopK, user.ID, input.TabID)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Error retrieving memories"})
        return
    }

    docs, err := ch.RAGService.Search(user.ID, input.TabID, input.Message, ch.TopK)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Error retrieving documents"})
        return
    }
	useReasoning := input.Reasoning != nil && *input.Reasoning
	var response string
    prompt := buildRAGPrompt(input.Message, memories, docs)
	if useReasoning {
		//reasoning path 
		reasoningPrompt := fmt.Sprintf( "You are a reasoning model. Analyze the context and produce a structured reasoning plan.\n\n%s", prompt, ) 
		//TODO add a specific model for reasoning to will be another interface
		reasoningOutput, err := ch.LLMService.GenerateResponse(reasoningPrompt) 
		if err != nil { 
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error generating reasoning"}) 
			return 
		} 
		finalPrompt := fmt.Sprintf( "Here is the reasoning:\n%s\n\nNow produce the final answer for the user.", reasoningOutput, ) 
		response, err = ch.LLMService.GenerateResponse(finalPrompt) 
		if err != nil { 
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error generating response"}) 
			return 
		} 
	} else {
		//direct answer path 
		response, err = ch.LLMService.GenerateResponse(prompt) 
		if err != nil { 
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error generating response"}) 
			return 
		} 
	}

    if err := ch.MemoryService.StoreMemory(
        fmt.Sprintf("Q: %s A: %s", input.Message, response),
        queryEmbedding,
        user.ID,
        input.TabID,
    ); err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Error storing memory"})
        return
    }

    c.JSON(http.StatusOK, gin.H{"response": response})
}

func buildRAGPrompt(userInput string, memories []models.Memory, docs []models.Document) string {
    var sb strings.Builder

    sb.WriteString("Relevant Document Context:\n")
    for _, d := range docs {
		//adding file name to context
		sb.WriteString("- File: ")
        sb.WriteString(d.Source) 
        sb.WriteString("\nContent: ")
        sb.WriteString(d.Content)
        sb.WriteString("\n")
    }

    sb.WriteString("\nRelevant Chat Memory:\n")
    for _, m := range memories {
        sb.WriteString("- ")
        sb.WriteString(m.Text)
        sb.WriteString("\n")
    }

    sb.WriteString("\nUser Question:\n")
    sb.WriteString(userInput)

    return sb.String()
}
