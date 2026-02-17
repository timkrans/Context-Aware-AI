package services

type LLMService interface {
    GenerateResponse(prompt string) (string, error)
}
