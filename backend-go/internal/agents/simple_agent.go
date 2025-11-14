package agents

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/Vova4o/bootCamp2025CaseSber/backend/internal/models"
	"github.com/Vova4o/bootCamp2025CaseSber/backend/internal/tools"
	"github.com/Vova4o/bootCamp2025CaseSber/backend/internal/utils"
)

type SimpleAgent struct {
	searchClient *tools.SearchClient
	llmClient    *tools.LLMClient
}

func NewSimpleAgent(searchClient *tools.SearchClient, llmClient *tools.LLMClient) *SimpleAgent {
	return &SimpleAgent{
		searchClient: searchClient,
		llmClient:    llmClient,
	}
}

func (a *SimpleAgent) Process(ctx context.Context, query string) (*models.SearchResponse, error) {
	return a.ProcessWithContext(ctx, query, nil)
}

func (a *SimpleAgent) ProcessWithContext(
	ctx context.Context,
	query string,
	conversationHistory []models.Message,
) (*models.SearchResponse, error) {
	log.Printf("Simple mode processing: %s (with context: %v)", query, len(conversationHistory) > 0)

	searchQuery := query

	// Step 1: Enhance query with context if available
	if len(conversationHistory) > 0 {
		var contextPrompt strings.Builder
		contextPrompt.WriteString("Предыдущая беседа:\n")
		start := len(conversationHistory) - 4
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

		enhanced, err := a.llmClient.Complete(ctx, enhancePrompt, 0.3, 150)
		if err == nil && enhanced != "" {
			searchQuery = enhanced
			log.Printf("Enhanced query: %s", searchQuery)
		}
	}

	// Step 2: Search for information
	searchResults, err := a.searchClient.Search(ctx, searchQuery, 5, false)
	if err != nil {
		return nil, fmt.Errorf("search failed: %w", err)
	}

	if len(searchResults.Results) == 0 {
		return &models.SearchResponse{
			Query:       query,
			Mode:        "simple",
			Answer:      "Не удалось найти релевантную информацию по вашему запросу.",
			Sources:     []models.Source{},
			ContextUsed: len(conversationHistory) > 0,
		}, nil
	}

	// Step 3: Format search results for LLM
	var sourcesContext strings.Builder
	sourcesContext.WriteString("Найденная информация:\n\n")
	for i, result := range searchResults.Results {
		content := utils.SanitizeUTF8(result.Content)
		sourcesContext.WriteString(fmt.Sprintf("Источник %d (%s):\n%s\n\n",
			i+1, result.Title, content))
	}

	// Step 4: Build LLM prompt
	var promptBuilder strings.Builder
	promptBuilder.WriteString("Ты поисковый ассистент. Дай краткий и точный ответ на вопрос пользователя на основе найденной информации.\n\n")

	if len(conversationHistory) > 0 {
		promptBuilder.WriteString("Контекст диалога:\n")
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
	promptBuilder.WriteString(sourcesContext.String())
	promptBuilder.WriteString("Ответ:")

	// Step 5: Generate answer using LLM
	answer, err := a.llmClient.Complete(ctx, promptBuilder.String(), 0.7, 500)
	if err != nil {
		return nil, fmt.Errorf("LLM completion failed: %w", err)
	}

	// Step 6: Format sources with UTF-8 safety
	sources := make([]models.Source, 0, len(searchResults.Results))
	for _, result := range searchResults.Results {
		snippet := utils.SanitizeUTF8(result.Snippet)
		if len(snippet) > 200 {
			snippet = utils.TruncateUTF8WithEllipsis(snippet, 200)
		}
		
		sources = append(sources, models.Source{
			Title:       utils.SanitizeUTF8(result.Title),
			URL:         result.URL,
			Snippet:     snippet,
			Credibility: result.Score,
		})
	}

	return &models.SearchResponse{
		Query:       query,
		Mode:        "simple",
		Answer:      answer,
		Sources:     sources,
		ContextUsed: len(conversationHistory) > 0,
	}, nil
}