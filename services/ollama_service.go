package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

type OllamaService struct {
	BaseURL        string
	GenerateModel  string
	EmbeddingModel string
}

type EmbeddingResponse struct {
	Embedding []float64 `json:"embedding"`
}

type GenerateResponse struct {
	Response string `json:"response"`
}

func (os *OllamaService) GetEmbedding(text string) ([]float64, error) {
	url := fmt.Sprintf("%s/api/embeddings", os.BaseURL)
	payload := fmt.Sprintf(`{"model": "%s", "prompt": %q}`, os.EmbeddingModel, text)
	req, _ := http.NewRequest("POST", url, bytes.NewBuffer([]byte(payload)))
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)
	var r EmbeddingResponse
	err = json.Unmarshal(body, &r)
	if err != nil {
		return nil, err
	}
	return r.Embedding, nil
}

func (os *OllamaService) GenerateResponse(prompt string) (string, error) {
	url := fmt.Sprintf("%s/api/generate", os.BaseURL)
	payload := fmt.Sprintf(`{"model":"%s","prompt":%q,"stream":false}`, os.GenerateModel, prompt)
	req, _ := http.NewRequest("POST", url, bytes.NewBuffer([]byte(payload)))
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)
	var r GenerateResponse
	json.Unmarshal(body, &r)
	return r.Response, nil
}
