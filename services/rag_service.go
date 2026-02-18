package services

import (
    "bytes"
    "context-aware-ai/models"
    "encoding/gob"
    "math"
    "sort"

    "gorm.io/gorm"
)

type RAGService struct {
    DB            *gorm.DB
    OllamaService *OllamaService
}

func encodeEmbedding(vec []float64) []byte {
    var buf bytes.Buffer
    _ = gob.NewEncoder(&buf).Encode(vec)
    return buf.Bytes()
}

func decodeEmbedding(b []byte) []float64 {
    var vec []float64
    _ = gob.NewDecoder(bytes.NewReader(b)).Decode(&vec)
    return vec
}

func cosine(a, b []float64) float64 {
    var dot, na, nb float64
    for i := range a {
        dot += a[i] * b[i]
        na += a[i] * a[i]
        nb += b[i] * b[i]
    }
    if na == 0 || nb == 0 {
        return 0
    }
    return dot / (math.Sqrt(na) * math.Sqrt(nb))
}

func (r *RAGService) DeleteDocumentsByTabID(userID, tabID uint) error {
    return r.DB.Where("user_id = ? AND tab_id = ?", userID, tabID).Delete(&models.Document{}).Error
}


func (r *RAGService) IndexChunk(userID, tabID uint, source, content string) error {
    emb, err := r.OllamaService.GetEmbedding(content)
    if err != nil {
        return err
    }

    doc := models.Document{
        UserID:    userID,
        TabID:     tabID,
        Source:    source,
        Content:   content,
        Embedding: encodeEmbedding(emb),
    }

    return r.DB.Create(&doc).Error
}

func (r *RAGService) Search(userID, tabID uint, query string, topK int) ([]models.Document, error) {
    qEmb, err := r.OllamaService.GetEmbedding(query)
    if err != nil {
        return nil, err
    }

    var docs []models.Document
    if err := r.DB.Where("user_id = ? AND tab_id = ?", userID, tabID).Find(&docs).Error; err != nil {
        return nil, err
    }

    type scored struct {
        Doc   models.Document
        Score float64
    }

    results := make([]scored, 0, len(docs))
    for _, d := range docs {
        emb := decodeEmbedding(d.Embedding)
        score := cosine(qEmb, emb)
        results = append(results, scored{Doc: d, Score: score})
    }

    sort.Slice(results, func(i, j int) bool {
        return results[i].Score > results[j].Score
    })

    if len(results) > topK {
        results = results[:topK]
    }

    final := make([]models.Document, len(results))
    for i, r := range results {
        final[i] = r.Doc
    }
    return final, nil
}
