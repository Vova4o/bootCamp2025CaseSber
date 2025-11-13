package tools

import (
	"context"
	"fmt"

	"github.com/Vova4o/bootCamp2025CaseSber/backend/internal/config"
	openai "github.com/sashabaranov/go-openai"
)

type LLMClient struct {
	cfg    *config.Config
	client *openai.Client
}

func NewLLMClient(cfg *config.Config) *LLMClient {
	var client *openai.Client

	// Use OpenAI by default
	if cfg.OpenAIKey != "" {
		client = openai.NewClient(cfg.OpenAIKey)
	} else if cfg.QwenAPIURL != "" {
		// For Qwen or other OpenAI-compatible APIs
		clientConfig := openai.DefaultConfig(cfg.OpenAIKey)
		clientConfig.BaseURL = cfg.QwenAPIURL
		client = openai.NewClientWithConfig(clientConfig)
	}

	return &LLMClient{
		cfg:    cfg,
		client: client,
	}
}

func (l *LLMClient) Complete(ctx context.Context, prompt string, temperature float32, maxTokens int) (string, error) {
	if l.client == nil {
		return "", fmt.Errorf("LLM client not initialized")
	}

	req := openai.ChatCompletionRequest{
		Model: l.cfg.OpenAIModel,
		Messages: []openai.ChatCompletionMessage{
			{
				Role:    openai.ChatMessageRoleUser,
				Content: prompt,
			},
		},
		Temperature: temperature,
		MaxTokens:   maxTokens,
	}

	resp, err := l.client.CreateChatCompletion(ctx, req)
	if err != nil {
		return "", fmt.Errorf("chat completion failed: %w", err)
	}

	if len(resp.Choices) == 0 {
		return "", fmt.Errorf("no response from LLM")
	}

	return resp.Choices[0].Message.Content, nil
}

func (l *LLMClient) ChatCompletion(
	ctx context.Context,
	messages []map[string]string,
	temperature float32,
	maxTokens int,
) (string, error) {
	if l.client == nil {
		return "", fmt.Errorf("LLM client not initialized")
	}

	var chatMessages []openai.ChatCompletionMessage
	for _, msg := range messages {
		role := msg["role"]
		content := msg["content"]

		var msgRole string
		switch role {
		case "system":
			msgRole = openai.ChatMessageRoleSystem
		case "assistant":
			msgRole = openai.ChatMessageRoleAssistant
		default:
			msgRole = openai.ChatMessageRoleUser
		}

		chatMessages = append(chatMessages, openai.ChatCompletionMessage{
			Role:    msgRole,
			Content: content,
		})
	}

	req := openai.ChatCompletionRequest{
		Model:       l.cfg.OpenAIModel,
		Messages:    chatMessages,
		Temperature: temperature,
		MaxTokens:   maxTokens,
	}

	resp, err := l.client.CreateChatCompletion(ctx, req)
	if err != nil {
		return "", fmt.Errorf("chat completion failed: %w", err)
	}

	if len(resp.Choices) == 0 {
		return "", fmt.Errorf("no response from LLM")
	}

	return resp.Choices[0].Message.Content, nil
}
