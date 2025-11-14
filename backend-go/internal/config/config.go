package config

import (
	"os"
	"strconv"
	"strings"
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

	// CORS
	CORSOrigins []string
}

func LoadConfig() *Config {
	debug, _ := strconv.ParseBool(getEnv("DEBUG", "true"))

	// Parse CORS origins
	corsOrigins := getEnv("CORS_ORIGINS", "http://localhost:3000")
	origins := strings.Split(corsOrigins, ",")
	// Trim whitespace from each origin
	for i, origin := range origins {
		origins[i] = strings.TrimSpace(origin)
	}

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

		CORSOrigins: origins,
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}