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

type FinanceAgent struct {
	financeScraper *scrapers.FinanceScraper
	llmClient      *tools.LLMClient
	reranker       *tools.BM25Reranker
}

func NewFinanceAgent(llmClient *tools.LLMClient) *FinanceAgent {
	return &FinanceAgent{
		financeScraper: scrapers.NewFinanceScraper(),
		llmClient:      llmClient,
		reranker:       tools.NewBM25Reranker(),
	}
}

func (a *FinanceAgent) Process(ctx context.Context, query string) (*models.SearchResponse, error) {
	return a.ProcessWithContext(ctx, query, nil)
}

func (a *FinanceAgent) ProcessWithContext(
	ctx context.Context,
	query string,
	conversationHistory []models.Message,
) (*models.SearchResponse, error) {
	log.Printf("Pro Finance mode processing: %s", query)

	reasoningSteps := []string{"💰 Запущен режим Finance - анализ финансовых данных"}

	searchQuery := query
	if len(conversationHistory) > 0 {
		reasoningSteps = append(reasoningSteps, "Адаптирую запрос с учетом контекста...")
		enhanced, err := a.enhanceQueryWithContext(ctx, query, conversationHistory)
		if err == nil && enhanced != "" {
			searchQuery = enhanced
		}
	}

	reasoningSteps = append(reasoningSteps, "Ищу финансовые данные в Yahoo Finance, Investing.com, MarketWatch...")

	allResults := make([]models.TavilyResult, 0)

	// Yahoo Finance
	yahooResults, err := a.financeScraper.SearchYahooFinance(ctx, searchQuery, 5)
	if err != nil {
		log.Printf("Yahoo Finance search failed: %v", err)
	} else {
		allResults = append(allResults, yahooResults...)
		reasoningSteps = append(reasoningSteps, fmt.Sprintf("✓ Yahoo Finance: %d новостей", len(yahooResults)))
	}

	// Investing.com
	investingResults, err := a.financeScraper.SearchInvestingCom(ctx, searchQuery, 5)
	if err != nil {
		log.Printf("Investing.com search failed: %v", err)
	} else {
		allResults = append(allResults, investingResults...)
		reasoningSteps = append(reasoningSteps, fmt.Sprintf("✓ Investing.com: %d результатов", len(investingResults)))
	}

	// MarketWatch
	marketwatchResults, err := a.financeScraper.SearchMarketWatch(ctx, searchQuery, 5)
	if err != nil {
		log.Printf("MarketWatch search failed: %v", err)
	} else {
		allResults = append(allResults, marketwatchResults...)
		reasoningSteps = append(reasoningSteps, fmt.Sprintf("✓ MarketWatch: %d статей", len(marketwatchResults)))
	}

	if len(allResults) == 0 {
		return &models.SearchResponse{
			Query:     query,
			Mode:      "pro-finance",
			Answer:    "Не удалось найти финансовую информацию по вашему запросу.",
			Sources:   []models.Source{},
			Reasoning: strings.Join(reasoningSteps, "\n"),
		}, nil
	}

	reasoningSteps = append(reasoningSteps, fmt.Sprintf("Собрано %d финансовых источников", len(allResults)))

	// Rerank
	allResults = a.reranker.Rerank(searchQuery, allResults)

	if len(allResults) > 10 {
		allResults = allResults[:10]
	}

	reasoningSteps = append(reasoningSteps, "Анализирую финансовые данные...")

	// Build LLM prompt
	var promptBuilder strings.Builder
	promptBuilder.WriteString(`Ты финансовый аналитик. Проанализируй финансовые данные и новости.

Твоя задача:
1. Дать объективный финансовый анализ
2. Указать ключевые факты и цифры
3. Отметить риски и возможности
4. Основываться только на проверенных источниках

⚠️ Важно: Это не финансовый совет. Пользователь должен провести собственное исследование.

`)

	if len(conversationHistory) > 0 {
		promptBuilder.WriteString("\nКонтекст диалога:\n")
		for _, msg := range conversationHistory[max(0, len(conversationHistory)-4):] {
			promptBuilder.WriteString(fmt.Sprintf("%s: %s\n", msg.Role, msg.Content))
		}
		promptBuilder.WriteString("\n")
	}

	promptBuilder.WriteString(fmt.Sprintf("Вопрос: %s\n\n", query))
	promptBuilder.WriteString("Финансовые источники:\n\n")

	for i, result := range allResults {
		if i >= 8 {
			break
		}
		content := result.Content
		if len(content) > 500 {
			content = content[:500]
		}
		promptBuilder.WriteString(fmt.Sprintf("Источник %d: %s\n%s\n\n", i+1, result.Title, content))
	}

	promptBuilder.WriteString("\nФинансовый анализ:")

	answer, err := a.llmClient.Complete(ctx, promptBuilder.String(), 0.6, 1000)
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
		Mode:        "pro-finance",
		Answer:      answer,
		Sources:     sources,
		Reasoning:   strings.Join(reasoningSteps, "\n"),
		ContextUsed: len(conversationHistory) > 0,
	}, nil
}

func (a *FinanceAgent) enhanceQueryWithContext(
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

Перефразируй текущий вопрос для поиска финансовой информации. Улучшенный запрос:`, contextPrompt.String(), query)

	return a.llmClient.Complete(ctx, enhancePrompt, 0.3, 150)
}