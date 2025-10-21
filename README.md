# auth-provider

## Claude AI Integration

This project includes integration with Anthropic's Claude AI. To use it:

1. Obtain an API key from [Anthropic](https://www.anthropic.com/)
2. Set the `ANTHROPIC_API_KEY` environment variable in your `.env` file
3. Inject the ClaudeClient into your services via dependency injection

### Example Usage

```go
package main

import (
    "context"
    "github.com/Oidiral/auth-provider/internal/ai"
    "github.com/sirupsen/logrus"
)

func main() {
    logger := logrus.New()
    
    // Create Claude client (will read ANTHROPIC_API_KEY from environment)
    claudeClient := ai.NewClaudeClient("", logger)
    
    // Send a prompt
    ctx := context.Background()
    response, err := claudeClient.SendPrompt(ctx, "claude-3-opus-20240229", "Hello, Claude!")
    if err != nil {
        logger.WithError(err).Error("Failed to get Claude response")
        return
    }
    
    logger.Info("Claude response: ", response)
}
```

### Configuration

See `.env.example` for all environment variable configurations, including `ANTHROPIC_API_KEY`.
