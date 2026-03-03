package review

import (
	"context"
	"fmt"
	"os"
	"sync"

	"google.golang.org/genai"
)

var (
	geminiOnce   sync.Once
	geminiClient *genai.Client
	geminiErr    error
)

// initClient creates the Gemini client exactly once (thread-safe).
func initClient(ctx context.Context) (*genai.Client, error) {
	geminiOnce.Do(func() {
		apiKey := os.Getenv("GOOGLE_API_KEY")
		if apiKey == "" {
			geminiErr = fmt.Errorf("GOOGLE_API_KEY not set")
			return
		}
		geminiClient, geminiErr = genai.NewClient(ctx, &genai.ClientConfig{
			APIKey:  apiKey,
			Backend: genai.BackendGeminiAPI,
		})
	})
	return geminiClient, geminiErr
}

// callGemini sends a prompt to the Gemini API and returns the response text.
// The client is created once and reused across all calls.
func callGemini(ctx context.Context, model, prompt string) (string, error) {
	client, err := initClient(ctx)
	if err != nil {
		return "", err
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
