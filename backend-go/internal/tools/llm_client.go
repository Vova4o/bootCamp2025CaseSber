package tools

import (
	"context"
	"fmt"
	"log"
	"strings"

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

// supportsCustomParams checks if model supports custom temperature and max_tokens
func (l *LLMClient) supportsCustomParams() bool {
	model := strings.ToLower(l.cfg.OpenAIModel)
	// o1 models and some newer GPT-4 variants don't support custom params
	if strings.Contains(model, "o1") ||
		strings.Contains(model, "o1-preview") ||
		strings.Contains(model, "o1-mini") {
		return false
	}
	return true
}

// isGPT4Model checks if model is GPT-4 or newer
func (l *LLMClient) isGPT4Model() bool {
	model := strings.ToLower(l.cfg.OpenAIModel)
	return strings.Contains(model, "gpt-4") || strings.Contains(model, "o1")
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
	}

	// Only set custom parameters for models that support them
	if l.supportsCustomParams() {
		req.Temperature = temperature
		if !l.isGPT4Model() {
			req.MaxTokens = maxTokens
		}
	}
	// For models that don't support custom params, use defaults (temperature=1, no max_tokens)

	resp, err := l.client.CreateChatCompletion(ctx, req)
	if err != nil {
		// Retry with default parameters if error is related to unsupported params
		if strings.Contains(err.Error(), "temperature") ||
			strings.Contains(err.Error(), "max_tokens") ||
			strings.Contains(err.Error(), "max_completion_tokens") {
			log.Printf("⚠️  Retrying with default parameters (temperature=1, no max_tokens)")
			
			req.Temperature = 1.0
			req.MaxTokens = 0
			
			resp, err = l.client.CreateChatCompletion(ctx, req)
			if err != nil {
				return "", fmt.Errorf("chat completion failed: %w", err)
			}
		} else {
			return "", fmt.Errorf("chat completion failed: %w", err)
		}
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
		Model:    l.cfg.OpenAIModel,
		Messages: chatMessages,
	}

	// Only set custom parameters for models that support them
	if l.supportsCustomParams() {
		req.Temperature = temperature
		if !l.isGPT4Model() {
			req.MaxTokens = maxTokens
		}
	}

	resp, err := l.client.CreateChatCompletion(ctx, req)
	if err != nil {
		// Retry with default parameters if error is related to unsupported params
		if strings.Contains(err.Error(), "temperature") ||
			strings.Contains(err.Error(), "max_tokens") ||
			strings.Contains(err.Error(), "max_completion_tokens") {
			log.Printf("⚠️  Retrying with default parameters (temperature=1, no max_tokens)")
			
			req.Temperature = 1.0
			req.MaxTokens = 0
			
			resp, err = l.client.CreateChatCompletion(ctx, req)
			if err != nil {
				return "", fmt.Errorf("chat completion failed: %w", err)
			}
		} else {
			return "", fmt.Errorf("chat completion failed: %w", err)
		}
	}

	if len(resp.Choices) == 0 {
		return "", fmt.Errorf("no response from LLM")
	}

	return resp.Choices[0].Message.Content, nil
}