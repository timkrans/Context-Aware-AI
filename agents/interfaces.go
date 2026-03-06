package agents

type LLM interface {
    GenerateResponse(prompt string) (string, error)
}

type Tool interface {
    Name() string
    Execute(args map[string]any) (string, error)
}

type Agent interface {
    Name() string
    Run(task string, context string) (string, error)
}

type Brain interface { 
	Decide(userPrompt string, context string) (*BrainDecision, error) 
}