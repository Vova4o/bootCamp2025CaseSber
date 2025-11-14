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

type AcademicAgent struct {
	academicScraper *scrapers.AcademicScraper
	llmClient       *tools.LLMClient
	reranker        *tools.BM25Reranker
}

func NewAcademicAgent(llmClient *tools.LLMClient) *AcademicAgent {
	return &AcademicAgent{
		academicScraper: scrapers.NewAcademicScraper(),
		llmClient:       llmClient,
		reranker:        tools.NewBM25Reranker(),
	}
}

func (a *AcademicAgent) Process(ctx context.Context, query string) (*models.SearchResponse, error) {
	return a.ProcessWithContext(ctx, query, nil)
}

func (a *AcademicAgent) ProcessWithContext(
	ctx context.Context,
	query string,
	conversationHistory []models.Message,
) (*models.SearchResponse, error) {
	log.Printf("Pro Academic mode processing: %s", query)

	reasoningSteps := []string{"🎓 Запущен режим Academic - поиск научных источников"}

	searchQuery := query
	if len(conversationHistory) > 0 {
		reasoningSteps = append(reasoningSteps, "Адаптирую запрос с учетом контекста...")
		enhanced, err := a.enhanceQueryWithContext(ctx, query, conversationHistory)
		if err == nil && enhanced != "" {
			searchQuery = enhanced
		}
	}

	reasoningSteps = append(reasoningSteps, "Ищу научные статьи в arXiv и Google Scholar...")

	allResults := make([]models.TavilyResult, 0)

	// arXiv
	arxivResults, err := a.academicScraper.SearchArxiv(ctx, searchQuery, 5)
	if err != nil {
		log.Printf("arXiv search failed: %v", err)
	} else {
		allResults = append(allResults, arxivResults...)
		reasoningSteps = append(reasoningSteps, fmt.Sprintf("✓ arXiv: %d статей", len(arxivResults)))
	}

	// Google Scholar
	scholarResults, err := a.academicScraper.SearchGoogleScholar(ctx, searchQuery, 5)
	if err != nil {
		log.Printf("Scholar search failed: %v", err)
	} else {
		allResults = append(allResults, scholarResults...)
		reasoningSteps = append(reasoningSteps, fmt.Sprintf("✓ Google Scholar: %d статей", len(scholarResults)))
	}

	if len(allResults) == 0 {
		return &models.SearchResponse{
			Query:     query,
			Mode:      "pro-academic",
			Answer:    "Не удалось найти научные статьи по вашему запросу.",
			Sources:   []models.Source{},
			Reasoning: strings.Join(reasoningSteps, "\n"),
		}, nil
	}

	reasoningSteps = append(reasoningSteps, fmt.Sprintf("Собрано %d научных источников", len(allResults)))

	// Rerank
	allResults = a.reranker.Rerank(searchQuery, allResults)

	if len(allResults) > 10 {
		allResults = allResults[:10]
	}

	reasoningSteps = append(reasoningSteps, "Анализирую научные результаты...")

	// Build LLM prompt
	var promptBuilder strings.Builder
	promptBuilder.WriteString(`Ты научный ассистент. Проанализируй академические источники.

Твоя задача:
1. Дать научно обоснованный ответ
2. Ссылаться на конкретные исследования
3. Указать консенсус или противоречия в научном сообществе
4. Отметить ключевые выводы

`)

	if len(conversationHistory) > 0 {
		promptBuilder.WriteString("\nКонтекст диалога:\n")
		for _, msg := range conversationHistory[max(0, len(conversationHistory)-4):] {
			promptBuilder.WriteString(fmt.Sprintf("%s: %s\n", msg.Role, msg.Content))
		}
		promptBuilder.WriteString("\n")
	}

	promptBuilder.WriteString(fmt.Sprintf("Вопрос: %s\n\n", query))
	promptBuilder.WriteString("Научные источники:\n\n")

	for i, result := range allResults {
		if i >= 8 {
			break
		}
		content := result.Content
		if len(content) > 600 {
			content = content[:600]
		}
		promptBuilder.WriteString(fmt.Sprintf("Источник %d: %s\n%s\n\n", i+1, result.Title, content))
	}

	promptBuilder.WriteString("\nНаучный анализ:")

	answer, err := a.llmClient.Complete(ctx, promptBuilder.String(), 0.6, 1200)
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
		Mode:        "pro-academic",
		Answer:      answer,
		Sources:     sources,
		Reasoning:   strings.Join(reasoningSteps, "\n"),
		ContextUsed: len(conversationHistory) > 0,
	}, nil
}

func (a *AcademicAgent) enhanceQueryWithContext(
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

Перефразируй текущий вопрос для поиска научных статей (более формально). Улучшенный запрос:`, contextPrompt.String(), query)

	return a.llmClient.Complete(ctx, enhancePrompt, 0.3, 150)
}