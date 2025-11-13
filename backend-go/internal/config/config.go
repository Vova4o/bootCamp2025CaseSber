package config

import (
	"os"
	"strconv"
)

type Config struct {
	// Server
	Port  string
	Debug bool

	// Database
	DatabaseURL string

	// Redis
	RedisURL string

	// LLM
	OpenAIKey    string
	OpenAIModel  string
	AnthropicKey string
	QwenAPIURL   string
	QwenModel    string

	// Social Media APIs
	RedditClientID     string
	RedditClientSecret string
	TwitterBearerToken string
	VKAccessToken      string
}

func LoadConfig() *Config {
	debug, _ := strconv.ParseBool(getEnv("DEBUG", "true"))

	return &Config{
		Port:  getEnv("PORT", "8000"),
		Debug: debug,

		DatabaseURL: getEnv("DATABASE_URL", "sqlite://research_pro.db"),
		RedisURL:    getEnv("REDIS_URL", "redis://localhost:6379"),

		OpenAIKey:    getEnv("OPENAI_API_KEY", ""),
		OpenAIModel:  getEnv("OPENAI_MODEL", "gpt-4"),
		AnthropicKey: getEnv("ANTHROPIC_API_KEY", ""),
		QwenAPIURL:   getEnv("QWEN_API_URL", ""),
		QwenModel:    getEnv("QWEN_MODEL", "qwen-turbo"),

		RedditClientID:     getEnv("REDDIT_CLIENT_ID", ""),
		RedditClientSecret: getEnv("REDDIT_CLIENT_SECRET", ""),
		TwitterBearerToken: getEnv("TWITTER_BEARER_TOKEN", ""),
		VKAccessToken:      getEnv("VK_ACCESS_TOKEN", ""),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
