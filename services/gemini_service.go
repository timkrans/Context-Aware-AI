package services

import (
    "bytes"
    "encoding/json"
    "fmt"
    "io/ioutil"
    "net/http"
)

type GeminiService struct {
    APIKey string
    Model  string
}

func (gs *GeminiService) GenerateResponse(prompt string) (string, error) {
    url := fmt.Sprintf(
        "https://generativelanguage.googleapis.com/v1beta/models/%s:generateContent?key=%s",
        gs.Model,
        gs.APIKey,
    )

    payload := map[string]interface{}{
        "contents": []map[string]interface{}{
            {
                "parts": []map[string]string{
                    {"text": prompt},
                },
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
        Candidates []struct {
            Content struct {
                Parts []struct {
                    Text string `json:"text"`
                } `json:"parts"`
            } `json:"content"`
        } `json:"candidates"`
    }

    if err := json.Unmarshal(body, &r); err != nil {
        return "", err
    }

    if len(r.Candidates) == 0 || len(r.Candidates[0].Content.Parts) == 0 {
        return "", fmt.Errorf("no response text found")
    }

    return r.Candidates[0].Content.Parts[0].Text, nil
}
