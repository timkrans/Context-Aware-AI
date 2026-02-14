package services

import (
	"encoding/json"
	"math"
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

func (s *MemoryService) RetrieveRelevant(queryEmbedding []float64, topK int, userID uint, tabID uint) ([]models.Memory, error) {
	memories, err := s.GetAllMemories(userID, tabID)
	if err != nil {
		return nil, err
	}

	type scoredMemory struct {
		Memory models.Memory
		Score  float64
	}

	var scored []scoredMemory

	for _, m := range memories {
		var emb []float64
		err := json.Unmarshal(m.Embedding, &emb)
		if err != nil {
			continue
		}
		score := cosineSimilarity(queryEmbedding, emb)
		scored = append(scored, scoredMemory{Memory: m, Score: score})
	}

	sort.Slice(scored, func(i, j int) bool {
		return scored[i].Score > scored[j].Score
	})

	var results []models.Memory
	for i := 0; i < topK && i < len(scored); i++ {
		results = append(results, scored[i].Memory)
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
