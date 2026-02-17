package services

import (
    "bytes"
    "encoding/json"
    "fmt"
    "io/ioutil"
    "net/http"
)

type ClaudeService struct {
    APIKey string
    Model  string
}

func (cs *ClaudeService) GenerateResponse(prompt string) (string, error) {
    url := "https://api.anthropic.com/v1/messages"

    payload := map[string]interface{}{
        "model":      cs.Model,
        "max_tokens": 150,
        "messages": []map[string]string{
            {
                "role":    "user",
                "content": prompt,
            },
        },
    }

    data, err := json.Marshal(payload)
    if err != nil {
        return "", err
    }

    req, err := http.NewRequest("POST", url, bytes.NewBuffer(data))
    if err != nil {
        return "", err
    }

    req.Header.Set("Content-Type", "application/json")
    req.Header.Set("x-api-key", cs.APIKey)
    req.Header.Set("anthropic-version", "2023-06-01")

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
        Content []struct {
            Text string `json:"text"`
        } `json:"content"`
    }

    if err := json.Unmarshal(body, &r); err != nil {
        return "", err
    }

    if len(r.Content) == 0 {
        return "", fmt.Errorf("no content returned")
    }

    return r.Content[0].Text, nil
}
