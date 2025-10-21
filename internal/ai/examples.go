// Package ai provides examples of integrating ClaudeClient into services
package ai

import (
	"context"

	"github.com/sirupsen/logrus"
)

// ExampleService demonstrates how to inject ClaudeClient into a service
type ExampleService struct {
	claudeClient *ClaudeClient
	logger       *logrus.Logger
}

// NewExampleService creates a new service with Claude AI integration
// This follows the dependency injection pattern used in the application
func NewExampleService(claudeClient *ClaudeClient, logger *logrus.Logger) *ExampleService {
	return &ExampleService{
		claudeClient: claudeClient,
		logger:       logger,
	}
}

// ProcessWithAI demonstrates using Claude AI in a service method
func (s *ExampleService) ProcessWithAI(ctx context.Context, userInput string) (string, error) {
	// Use Claude to process the input
	response, err := s.claudeClient.SendPrompt(ctx, "claude-3-opus-20240229", userInput)
	if err != nil {
		s.logger.WithError(err).Error("Failed to process with Claude")
		return "", err
	}

	s.logger.Info("Successfully processed input with Claude AI")
	return response, nil
}

// Example of how to wire up the service in app.go:
//
// func Run() {
//     cfg, err := config.Init()
//     if err != nil {
//         logger.Error(err)
//         return
//     }
//
//     // Initialize logger
//     log := logrus.New()
//
//     // Initialize Claude client with API key from config
//     claudeClient := ai.NewClaudeClient(cfg.AI.AnthropicAPIKey, log)
//
//     // Initialize your service with Claude client injected
//     exampleService := NewExampleService(claudeClient, log)
//
//     // Use the service...
// }
