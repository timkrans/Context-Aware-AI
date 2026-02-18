package handlers

import (
    "context-aware-ai/services"
    "io"
    "net/http"
    "strconv"
	"strings"
    "github.com/gin-gonic/gin"
)

type FileHandler struct {
    RAGService   *services.RAGService
    ChatHandler  *ChatHandler//for authentication and gettabs
}

func (h *FileHandler) SetupRoutes(router *gin.Engine) {
    router.POST("/upload", h.Upload)
}

func (h *FileHandler) Upload(c *gin.Context) {
    user, err := h.ChatHandler.Authenticate(c)
    if err != nil {
        return
    }

    tabIDStr := c.PostForm("tab_id")
	tabID, err := strconv.Atoi(tabIDStr)
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid tab ID"})
        return
    }

    tabs, err := h.ChatHandler.TabService.GetTabs(user.ID)
    if err != nil || tabID < 1 || tabID > len(tabs) {
        c.JSON(http.StatusNotFound, gin.H{"error": "Tab not found"})
        return
    }

    tab := tabs[tabID-1]

    file, err := c.FormFile("file")
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "file required"})
        return
    }

    f, _ := file.Open()
    defer f.Close()
    data, _ := io.ReadAll(f)

    chunks := chunkText(string(data), 300, 50)

    for _, ch := range chunks {
        if err := h.RAGService.IndexChunk(user.ID, tab.ID, file.Filename, ch); err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": "error indexing document"})
            return
        }
    }

    c.JSON(http.StatusOK, gin.H{"status": "indexed"})
}

func chunkText(text string, size, overlap int) []string {
    words := strings.Fields(text)
    var chunks []string
    for i := 0; i < len(words); {
        end := i + size
        if end > len(words) {
            end = len(words)
        }
        chunks = append(chunks, strings.Join(words[i:end], " "))
        if end == len(words) {
            break
        }
        i += size - overlap
    }
    return chunks
}
