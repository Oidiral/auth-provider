package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/sirupsen/logrus"
)

type ClaudeClient struct {
	apiKey  string
	baseURL string
	client  *http.Client
	log     *logrus.Logger
}

func NewClaudeClient(apiKey string, logger *logrus.Logger) *ClaudeClient {
	if apiKey == "" {
		apiKey = os.Getenv("ANTHROPIC_API_KEY")
	}
	if logger == nil {
		logger = logrus.New()
	}
	return &ClaudeClient{
		apiKey:  apiKey,
		baseURL: "https://api.anthropic.com/v1",
		client:  http.DefaultClient,
		log:     logger,
	}
}

// SendPrompt sends request to Anthropic Claude and returns raw response.
func (c *ClaudeClient) SendPrompt(ctx context.Context, model, input string) (string, error) {
	if c.apiKey == "" {
		return "", fmt.Errorf("missing Anthropic API key")
	}

	reqBody := map[string]interface{}{
		"model": model,
		"input": input,
	}
	b, _ := json.Marshal(reqBody)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/responses", bytes.NewReader(b))
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		c.log.WithError(err).Error("claude request failed")
		return "", err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 400 {
		c.log.WithField("status", resp.StatusCode).Error("claude returned error")
		return "", fmt.Errorf("claude error: %s", string(body))
	}

	return string(body), nil
}
