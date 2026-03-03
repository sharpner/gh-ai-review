package review

import (
	"context"
	"fmt"
	"os"

	"google.golang.org/genai"
)

// callGemini sends a prompt to the Gemini API and returns the response text.
func callGemini(ctx context.Context, model, prompt string) (string, error) {
	apiKey := os.Getenv("GOOGLE_API_KEY")
	if apiKey == "" {
		return "", fmt.Errorf("GOOGLE_API_KEY not set")
	}

	client, err := genai.NewClient(ctx, &genai.ClientConfig{
		APIKey:  apiKey,
		Backend: genai.BackendGeminiAPI,
	})
	if err != nil {
		return "", fmt.Errorf("create genai client: %w", err)
	}

	result, err := client.Models.GenerateContent(ctx, model, genai.Text(prompt), nil)
	if err != nil {
		return "", fmt.Errorf("generate content: %w", err)
	}

	text := result.Text()
	if text == "" {
		return "", fmt.Errorf("empty response from Gemini")
	}

	return text, nil
}
