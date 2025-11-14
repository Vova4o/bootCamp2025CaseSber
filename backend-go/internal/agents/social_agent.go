package agents

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/Vova4o/bootCamp2025CaseSber/backend/internal/models"
	"github.com/Vova4o/bootCamp2025CaseSber/backend/internal/scrapers"
	"github.com/Vova4o/bootCamp2025CaseSber/backend/internal/tools"
)

type SocialAgent struct {
	socialScraper *scrapers.SocialScraper
	llmClient     *tools.LLMClient
	reranker      *tools.BM25Reranker
}

func NewSocialAgent(llmClient *tools.LLMClient) *SocialAgent {
	return &SocialAgent{
		socialScraper: scrapers.NewSocialScraper(),
		llmClient:     llmClient,
		reranker:      tools.NewBM25Reranker(),
	}
}

func (a *SocialAgent) Process(ctx context.Context, query string) (*models.SearchResponse, error) {
	return a.ProcessWithContext(ctx, query, nil)
}

func (a *SocialAgent) ProcessWithContext(
	ctx context.Context,
	query string,
	conversationHistory []models.Message,
) (*models.SearchResponse, error) {
	log.Printf("Pro Social mode processing: %s", query)

	reasoningSteps := []string{"🗣️ Запущен режим Social - анализ мнений и дискуссий"}

	searchQuery := query
	if len(conversationHistory) > 0 {
		reasoningSteps = append(reasoningSteps, "Адаптирую запрос с учетом контекста...")
		enhanced, err := a.enhanceQueryWithContext(ctx, query, conversationHistory)
		if err == nil && enhanced != "" {
			searchQuery = enhanced
		}
	}

	// Параллельный поиск в социальных сетях
	reasoningSteps = append(reasoningSteps, "Ищу мнения в Reddit, Habr, Twitter...")

	allResults := make([]models.TavilyResult, 0)

	// Reddit
	redditResults, err := a.socialScraper.SearchReddit(ctx, searchQuery, 5)
	if err != nil {
		log.Printf("Reddit search failed: %v", err)
	} else {
		allResults = append(allResults, redditResults...)
		reasoningSteps = append(reasoningSteps, fmt.Sprintf("✓ Reddit: %d обсуждений", len(redditResults)))
	}

	// Habr
	habrResults, err := a.socialScraper.SearchHabr(ctx, searchQuery, 5)
	if err != nil {
		log.Printf("Habr search failed: %v", err)
	} else {
		allResults = append(allResults, habrResults...)
		reasoningSteps = append(reasoningSteps, fmt.Sprintf("✓ Habr: %d статей", len(habrResults)))
	}

	// Twitter
	twitterResults, err := a.socialScraper.SearchTwitter(ctx, searchQuery, 5)
	if err != nil {
		log.Printf("Twitter search failed: %v", err)
	} else {
		allResults = append(allResults, twitterResults...)
		reasoningSteps = append(reasoningSteps, fmt.Sprintf("✓ Twitter: %d твитов", len(twitterResults)))
	}

	if len(allResults) == 0 {
		return &models.SearchResponse{
			Query:     query,
			Mode:      "pro-social",
			Answer:    "Не удалось найти обсуждения по вашему запросу в социальных сетях.",
			Sources:   []models.Source{},
			Reasoning: strings.Join(reasoningSteps, "\n"),
		}, nil
	}

	reasoningSteps = append(reasoningSteps, fmt.Sprintf("Собрано %d источников, применяю reranking...", len(allResults)))

	// Rerank
	allResults = a.reranker.Rerank(searchQuery, allResults)

	// Take top 10
	if len(allResults) > 10 {
		allResults = allResults[:10]
	}

	// Analyze sentiment
	reasoningSteps = append(reasoningSteps, "Анализирую тональность и общее мнение...")

	// Build LLM prompt
	var promptBuilder strings.Builder
	promptBuilder.WriteString(`Ты аналитик социальных медиа. Проанализируй мнения из разных источников.

Твоя задача:
1. Обобщить основные мнения и точки зрения
2. Выявить консенсус или противоречия
3. Указать тональность (позитивная/негативная/нейтральная)
4. Отметить наиболее популярные аргументы

`)

	if len(conversationHistory) > 0 {
		promptBuilder.WriteString("\nКонтекст диалога:\n")
		for _, msg := range conversationHistory[max(0, len(conversationHistory)-4):] {
			promptBuilder.WriteString(fmt.Sprintf("%s: %s\n", msg.Role, msg.Content))
		}
		promptBuilder.WriteString("\n")
	}

	promptBuilder.WriteString(fmt.Sprintf("Вопрос: %s\n\n", query))
	promptBuilder.WriteString("Найденные мнения:\n\n")

	for i, result := range allResults {
		if i >= 8 {
			break
		}
		content := result.Content
		if len(content) > 500 {
			content = content[:500]
		}
		promptBuilder.WriteString(fmt.Sprintf("Источник %d (%s):\n%s\n\n", i+1, result.Title, content))
	}

	promptBuilder.WriteString("\nАнализ мнений:")

	reasoningSteps = append(reasoningSteps, "Формирую итоговый анализ...")

	answer, err := a.llmClient.Complete(ctx, promptBuilder.String(), 0.7, 1000)
	if err != nil {
		return nil, fmt.Errorf("LLM completion failed: %w", err)
	}

	// Format sources
	sources := make([]models.Source, 0)
	for i, result := range allResults {
		if i >= 8 {
			break
		}
		snippet := result.Content
		if len(snippet) > 200 {
			snippet = snippet[:200] + "..."
		}
		sources = append(sources, models.Source{
			Title:       result.Title,
			URL:         result.URL,
			Snippet:     snippet,
			Credibility: result.Score,
		})
	}

	return &models.SearchResponse{
		Query:       query,
		Mode:        "pro-social",
		Answer:      answer,
		Sources:     sources,
		Reasoning:   strings.Join(reasoningSteps, "\n"),
		ContextUsed: len(conversationHistory) > 0,
	}, nil
}

func (a *SocialAgent) enhanceQueryWithContext(
	ctx context.Context,
	query string,
	conversationHistory []models.Message,
) (string, error) {
	var contextPrompt strings.Builder
	contextPrompt.WriteString("Предыдущая беседа:\n")
	for _, msg := range conversationHistory[max(0, len(conversationHistory)-4):] {
		role := "Пользователь"
		if msg.Role == "assistant" {
			role = "Ассистент"
		}
		contextPrompt.WriteString(fmt.Sprintf("%s: %s\n", role, msg.Content))
	}

	enhancePrompt := fmt.Sprintf(`%s

Текущий вопрос: %s

Перефразируй текущий вопрос так, чтобы он был самодостаточным для поиска в социальных сетях. Улучшенный запрос:`, contextPrompt.String(), query)

	return a.llmClient.Complete(ctx, enhancePrompt, 0.3, 150)
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}