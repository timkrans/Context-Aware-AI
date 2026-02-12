package services

import (
	"encoding/json"
	"math"
	"sort"

	"gorm.io/gorm"
	"context-aware-ai/models"
)

type MemoryService struct {
	DB *gorm.DB
}

func NewMemoryService(db *gorm.DB) *MemoryService {
	db.AutoMigrate(&models.Memory{})
	return &MemoryService{DB: db}
}

func (s *MemoryService) StoreMemory(text string, embedding []float64) error {
	data, err := json.Marshal(embedding)
	if err != nil {
		return err
	}

	mem := models.Memory{
		Text:      text,
		Embedding: data,
	}

	return s.DB.Create(&mem).Error
}

func (s *MemoryService) GetAllMemories() ([]models.Memory, error) {
	var memories []models.Memory
	err := s.DB.Find(&memories).Error
	return memories, err
}

func (s *MemoryService) RetrieveRelevant(queryEmbedding []float64, topK int) ([]models.Memory, error) {
	memories, err := s.GetAllMemories()
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
