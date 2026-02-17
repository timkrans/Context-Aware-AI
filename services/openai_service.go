package services

import (
    "bytes"
    "encoding/json"
    "fmt"
    "io/ioutil"
    "net/http"
)

type OpenAIService struct {
    APIKey string
    Model  string
}

func (os *OpenAIService) GenerateResponse(prompt string) (string, error) {
    url := "https://api.openai.com/v1/chat/completions"

    payload := map[string]interface{}{
        "model": os.Model,
        "messages": []map[string]string{
            {
                "role":    "user",
                "content": prompt,
            },
        },
        "max_tokens": 150,
    }

    data, err := json.Marshal(payload)
    if err != nil {
        return "", err
    }

    req, err := http.NewRequest("POST", url, bytes.NewBuffer(data))
    if err != nil {
        return "", err
    }

    req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", os.APIKey))
    req.Header.Set("Content-Type", "application/json")

    resp, err := http.DefaultClient.Do(req)
    if err != nil {
        return "", err
    }
    defer resp.Body.Close()

    body, err := ioutil.ReadAll(resp.Body)
    if err != nil {
        return "", err
    }

    var r struct {
        Choices []struct {
            Message struct {
                Content string `json:"content"`
            } `json:"message"`
        } `json:"choices"`
    }

    if err := json.Unmarshal(body, &r); err != nil {
        return "", err
    }

    if len(r.Choices) == 0 {
        return "", fmt.Errorf("no choices returned")
    }

    return r.Choices[0].Message.Content, nil
}
