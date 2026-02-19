package services

import (
	"encoding/json"
	"math"
	"time"
	"sort"
	"context-aware-ai/models"
	"gorm.io/gorm"
)

type MemoryService struct {
	DB *gorm.DB
}

func NewMemoryService(db *gorm.DB) *MemoryService {
	db.AutoMigrate(&models.Memory{})
	return &MemoryService{DB: db}
}

func (s *MemoryService) StoreMemory(text string, embedding []float64, userID uint, tabID uint) error {
	data, err := json.Marshal(embedding)
	if err != nil {
		return err
	}

	mem := models.Memory{
		Text:      text,
		Embedding: data,
		UserID:    userID,
		TabID:     tabID,
	}

	return s.DB.Create(&mem).Error
}

func (s *MemoryService) GetAllMemories(userID uint, tabID uint) ([]models.Memory, error) {
	var memories []models.Memory
	err := s.DB.Where("user_id = ? AND tab_id = ?", userID, tabID).Find(&memories).Error
	return memories, err
}

func (s *MemoryService) RetrieveRelevant(queryEmbedding []float64,topK int, userID uint,tabID uint,) ([]models.Memory, error) {
    memories, err := s.GetAllMemories(userID, tabID)
    if err != nil {
        return nil, err
    }
    if len(memories) == 0 {
        return []models.Memory{}, nil
    }

    type scoredMemory struct {
        Memory models.Memory
        Score  float64
    }

    scored := make([]scoredMemory, 0, len(memories))
    newest := memories[0].CreatedAt
    oldest := memories[0].CreatedAt

    for _, m := range memories {
        if m.CreatedAt.After(newest) {
            newest = m.CreatedAt
        }
        if m.CreatedAt.Before(oldest) {
            oldest = m.CreatedAt
        }
    }

    timeRange := newest.Sub(oldest)
    if timeRange == 0 {
        timeRange = time.Second 
    }
	//weight of consine similarity
    alpha := 0.8

    for _, m := range memories {
        var emb []float64
        if err := json.Unmarshal(m.Embedding, &emb); err != nil {
            continue
        }

        cos := cosineSimilarity(queryEmbedding, emb)

        recency := float64(m.CreatedAt.Sub(oldest)) / float64(timeRange)

        finalScore := alpha*cos + (1-alpha)*recency

        scored = append(scored, scoredMemory{
            Memory: m,
            Score:  finalScore,
        })
    }

    sort.Slice(scored, func(i, j int) bool {
        return scored[i].Score > scored[j].Score
    })

    if topK > len(scored) {
        topK = len(scored)
    }

    results := make([]models.Memory, topK)
    for i := 0; i < topK; i++ {
        results[i] = scored[i].Memory
    }

    return results, nil
}


func cosineSimilarity(a, b []float64) float64 {
	if len(a) != len(b) {
		return 0
	}
	var dot, normA, normB float64
	for i := range a {
		dot += a[i] * b[i]
		normA += a[i] * a[i]
		normB += b[i] * b[i]
	}
	if normA == 0 || normB == 0 {
		return 0
	}
	return dot / (math.Sqrt(normA) * math.Sqrt(normB))
}

func (s *MemoryService) DeleteMemoriesByTabID(userID uint, tabID uint) error {
    err := s.DB.Where("user_id = ? AND tab_id = ?", userID, tabID).Delete(&models.Memory{}).Error
    return err
}
