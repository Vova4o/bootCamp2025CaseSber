package agents

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/Vova4o/bootCamp2025CaseSber/backend/internal/models"
	"github.com/Vova4o/bootCamp2025CaseSber/backend/internal/tools"
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
	log.Printf("Simple mode processing: %s", query)

	// Step 1: Search for information
	searchResults, err := a.searchClient.Search(ctx, query, 5, false)
	if err != nil {
		return nil, fmt.Errorf("search failed: %w", err)
	}

	if len(searchResults.Results) == 0 {
		return &models.SearchResponse{
			Query:   query,
			Mode:    "simple",
			Answer:  "Не удалось найти релевантную информацию по вашему запросу.",
			Sources: []models.Source{},
		}, nil
	}

	// Step 2: Format search results for LLM
	var sourcesContext strings.Builder
	sourcesContext.WriteString("Найденная информация:\n\n")
	for i, result := range searchResults.Results {
		sourcesContext.WriteString(fmt.Sprintf("Источник %d (%s):\n%s\n\n",
			i+1, result.Title, result.Content))
	}

	// Step 3: Generate answer using LLM
	prompt := fmt.Sprintf(`Ты поисковый ассистент. Дай краткий и точный ответ на вопрос пользователя на основе найденной информации.

Вопрос: %s

%s

Ответ:`, query, sourcesContext.String())

	answer, err := a.llmClient.Complete(ctx, prompt, 0.7, 500)
	if err != nil {
		return nil, fmt.Errorf("LLM completion failed: %w", err)
	}

	// Step 4: Format sources
	sources := make([]models.Source, 0, len(searchResults.Results))
	for _, result := range searchResults.Results {
		snippet := result.Snippet
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
		Query:   query,
		Mode:    "simple",
		Answer:  answer,
		Sources: sources,
	}, nil
}
