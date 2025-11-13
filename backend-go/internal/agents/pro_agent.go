package agents

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/Vova4o/bootCamp2025CaseSber/backend/internal/models"
	"github.com/Vova4o/bootCamp2025CaseSber/backend/internal/tools"
)

type ProAgent struct {
	searchClient *tools.SearchClient
	llmClient    *tools.LLMClient
}

func NewProAgent(searchClient *tools.SearchClient, llmClient *tools.LLMClient) *ProAgent {
	return &ProAgent{
		searchClient: searchClient,
		llmClient:    llmClient,
	}
}

func (a *ProAgent) Process(ctx context.Context, query string) (*models.SearchResponse, error) {
	return a.ProcessWithContext(ctx, query, nil)
}

func (a *ProAgent) ProcessWithContext(
	ctx context.Context,
	query string,
	conversationHistory []models.Message,
) (*models.SearchResponse, error) {
	log.Printf("Pro mode processing: %s (with context: %v)", query, len(conversationHistory) > 0)

	reasoningSteps := []string{}
	searchQuery := query

	// Step 1: Enhance query with context if available
	if len(conversationHistory) > 0 {
		reasoningSteps = append(reasoningSteps, "🔍 Анализирую контекст предыдущего диалога...")

		var contextPrompt strings.Builder
		contextPrompt.WriteString("Предыдущая беседа:\n")
		// Take last 6 messages (3 pairs)
		start := len(conversationHistory) - 6
		if start < 0 {
			start = 0
		}
		for _, msg := range conversationHistory[start:] {
			role := "Пользователь"
			if msg.Role == "assistant" {
				role = "Ассистент"
			}
			contextPrompt.WriteString(fmt.Sprintf("\n%s: %s\n", role, msg.Content))
		}

		enhancePrompt := fmt.Sprintf(`%s

Текущий вопрос: %s

Перефразируй текущий вопрос так, чтобы он был самодостаточным и включал важную информацию из контекста. Улучшенный поисковый запрос:`, contextPrompt.String(), query)

		enhanced, err := a.llmClient.Complete(ctx, enhancePrompt, 0.3, 200)
		if err == nil && enhanced != "" {
			searchQuery = enhanced
			reasoningSteps = append(reasoningSteps, fmt.Sprintf("✨ Улучшенный запрос: %s", searchQuery))
		}
	} else {
		reasoningSteps = append(reasoningSteps, "📝 Обрабатываю первый запрос без контекста")
	}

	// Step 2: Search for information
	reasoningSteps = append(reasoningSteps, fmt.Sprintf("🔎 Ищу информацию по запросу: %s", searchQuery))

	searchResults, err := a.searchClient.Search(ctx, searchQuery, 10, true)
	if err != nil {
		return nil, fmt.Errorf("search failed: %w", err)
	}

	if len(searchResults.Results) == 0 {
		return &models.SearchResponse{
			Query:     query,
			Mode:      "pro",
			Answer:    "Не удалось найти релевантную информацию по вашему запросу.",
			Sources:   []models.Source{},
			Reasoning: strings.Join(reasoningSteps, "\n"),
		}, nil
	}

	reasoningSteps = append(reasoningSteps, fmt.Sprintf("✅ Найдено %d источников", len(searchResults.Results)))

	// Step 3: Format sources for LLM
	var sourcesContext strings.Builder
	for i, result := range searchResults.Results {
		if i >= 5 {
			break
		}
		content := result.Content
		if result.RawContent != "" {
			content = result.RawContent
		}
		if len(content) > 1000 {
			content = content[:1000]
		}
		sourcesContext.WriteString(fmt.Sprintf("Источник %d (%s):\n%s\n\n", i+1, result.Title, content))
	}

	// Step 4: Build LLM prompt with context
	var promptBuilder strings.Builder
	promptBuilder.WriteString(`Ты исследовательский ассистент в режиме Pro.
Твоя задача - дать подробный, хорошо обоснованный ответ с учётом:
1. Контекста предыдущей беседы (если есть)
2. Найденных источников
3. Проверки фактов

Формат ответа:
- Прямой ответ на вопрос
- Подтверждение фактами из источников
- Если информация противоречива - укажи это

`)

	// Add conversation history
	if len(conversationHistory) > 0 {
		promptBuilder.WriteString("\nКонтекст диалога:\n")
		start := len(conversationHistory) - 4
		if start < 0 {
			start = 0
		}
		for _, msg := range conversationHistory[start:] {
			promptBuilder.WriteString(fmt.Sprintf("%s: %s\n", msg.Role, msg.Content))
		}
		promptBuilder.WriteString("\n")
	}

	promptBuilder.WriteString(fmt.Sprintf("Вопрос: %s\n\n", query))
	promptBuilder.WriteString("Найденная информация:\n")
	promptBuilder.WriteString(sourcesContext.String())
	promptBuilder.WriteString("\nОтвет:")

	reasoningSteps = append(reasoningSteps, "💡 Формирую ответ с учётом всех данных...")

	// Step 5: Generate answer
	answer, err := a.llmClient.Complete(ctx, promptBuilder.String(), 0.7, 1000)
	if err != nil {
		return nil, fmt.Errorf("LLM completion failed: %w", err)
	}

	// Step 6: Format sources
	sources := make([]models.Source, 0)
	for i, result := range searchResults.Results {
		if i >= 5 {
			break
		}
		snippet := result.Snippet
		if len(snippet) > 200 {
			snippet = snippet[:200] + "..."
		}
		sources = append(sources, models.Source{
			Title:       result.Title,
			URL:         result.URL,
			Snippet:     snippet,
			Credibility: 0.85, // TODO: Implement real credibility scoring
		})
	}

	return &models.SearchResponse{
		Query:       query,
		Mode:        "pro",
		Answer:      answer,
		Sources:     sources,
		Reasoning:   strings.Join(reasoningSteps, "\n"),
		ContextUsed: len(conversationHistory) > 0,
	}, nil
}
