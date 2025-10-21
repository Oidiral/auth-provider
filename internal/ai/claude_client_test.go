package ai_test

import (
	"context"
	"testing"

	"github.com/Oidiral/auth-provider/internal/ai"
	"github.com/sirupsen/logrus"
)

// Example demonstrates how to create and use ClaudeClient
func ExampleNewClaudeClient() {
	logger := logrus.New()

	// Create Claude client with explicit API key
	client := ai.NewClaudeClient("your-api-key-here", logger)

	// Or rely on ANTHROPIC_API_KEY environment variable
	client = ai.NewClaudeClient("", logger)

	_ = client
}

// Example demonstrates sending a prompt to Claude
func ExampleClaudeClient_SendPrompt() {
	logger := logrus.New()
	client := ai.NewClaudeClient("", logger)

	ctx := context.Background()
	response, err := client.SendPrompt(ctx, "claude-3-opus-20240229", "Hello, Claude!")
	if err != nil {
		logger.WithError(err).Error("Failed to send prompt")
		return
	}

	logger.Info("Response: ", response)
}

// TestNewClaudeClient_WithAPIKey tests client creation with explicit API key
func TestNewClaudeClient_WithAPIKey(t *testing.T) {
	logger := logrus.New()
	apiKey := "test-api-key"

	client := ai.NewClaudeClient(apiKey, logger)
	if client == nil {
		t.Fatal("Expected non-nil client")
	}
}

// TestNewClaudeClient_WithoutAPIKey tests client creation without explicit API key
func TestNewClaudeClient_WithoutAPIKey(t *testing.T) {
	logger := logrus.New()

	client := ai.NewClaudeClient("", logger)
	if client == nil {
		t.Fatal("Expected non-nil client")
	}
}

// TestNewClaudeClient_WithNilLogger tests client creation with nil logger
func TestNewClaudeClient_WithNilLogger(t *testing.T) {
	client := ai.NewClaudeClient("test-api-key", nil)
	if client == nil {
		t.Fatal("Expected non-nil client even with nil logger")
	}
}

// TestSendPrompt_MissingAPIKey tests error handling when API key is missing
func TestSendPrompt_MissingAPIKey(t *testing.T) {
	logger := logrus.New()
	client := ai.NewClaudeClient("", logger)

	// Ensure no API key is set in environment for this test
	t.Setenv("ANTHROPIC_API_KEY", "")

	ctx := context.Background()
	_, err := client.SendPrompt(ctx, "claude-3-opus-20240229", "test")
	if err == nil {
		t.Fatal("Expected error when API key is missing")
	}
	if err.Error() != "missing Anthropic API key" {
		t.Fatalf("Expected 'missing Anthropic API key' error, got: %v", err)
	}
}
