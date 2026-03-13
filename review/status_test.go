package review

import (
	"context"
	"testing"
)

func TestStatusResult_OK_AllPass(t *testing.T) {
	r := StatusResult{
		Checks: []StatusCheck{
			{Name: "A", OK: true},
			{Name: "B", OK: true},
		},
	}
	if !r.OK() {
		t.Error("expected OK() = true when all checks pass")
	}
}

func TestStatusResult_OK_OneFails(t *testing.T) {
	r := StatusResult{
		Checks: []StatusCheck{
			{Name: "A", OK: true},
			{Name: "B", OK: false},
		},
	}
	if r.OK() {
		t.Error("expected OK() = false when one check fails")
	}
}

func TestStatusResult_OK_Empty(t *testing.T) {
	r := StatusResult{}
	if r.OK() {
		t.Error("expected OK() = false for empty checks")
	}
}

func TestCheckStatus_UnknownProvider(t *testing.T) {
	result := CheckStatus(context.Background(), "openai", "gpt-4")
	if result.OK() {
		t.Error("expected failure for unknown provider")
	}
	if len(result.Checks) != 1 {
		t.Fatalf("expected 1 check, got %d", len(result.Checks))
	}
	if result.Checks[0].OK {
		t.Error("expected check to fail")
	}
}

func TestCheckStatus_GeminiNoKey(t *testing.T) {
	t.Setenv("GOOGLE_API_KEY", "")

	result := CheckStatus(context.Background(), ProviderGemini, "gemini-2.5-flash")
	if result.OK() {
		t.Error("expected failure when GOOGLE_API_KEY is not set")
	}
	if len(result.Checks) < 1 {
		t.Fatal("expected at least 1 check")
	}
	if result.Checks[0].Name != "API Key" {
		t.Errorf("first check should be API Key, got %s", result.Checks[0].Name)
	}
	if result.Checks[0].OK {
		t.Error("API Key check should fail")
	}
}
