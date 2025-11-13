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
	searchClient       *tools.SearchClient
	llmClient          *tools.LLMClient
	reranker           *tools.BM25Reranker
	credibilityScorer  *tools.CredibilityScorer
}

func NewProAgent(searchClient *tools.SearchClient, llmClient *tools.LLMClient) *ProAgent {
	return &ProAgent{
		searchClient:      searchClient,
		llmClient:         llmClient,
		reranker:          tools.NewBM25Reranker(),
		credibilityScorer: tools.NewCredibilityScorer(),
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

	// Step 1: Enhance query with context
	if len(conversationHistory) > 0 {
		reasoningSteps = append(reasoningSteps, "🔍 Анализирую контекст предыдущего диалога...")

		var contextPrompt strings.Builder
		contextPrompt.WriteString("Предыдущая беседа:\n")
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

	// Step 2: Определяем, нужен ли multi-hop
	needsMultiHop := a.detectMultiHop(query)
	
	var allResults []models.TavilyResult
	
	if needsMultiHop {
		reasoningSteps = append(reasoningSteps, "🔗 Обнаружен сложный вопрос - применяю multi-hop reasoning")
		
		// Разбиваем на подвопросы
		subQueries := a.generateSubQueries(ctx, searchQuery)
		reasoningSteps = append(reasoningSteps, fmt.Sprintf("📊 Разбил на %d подвопроса", len(subQueries)))
		
		// Ищем ответы на каждый подвопрос
		for i, subQuery := range subQueries {
			reasoningSteps = append(reasoningSteps, fmt.Sprintf("🔎 Подзапрос %d: %s", i+1, subQuery))
			
			results, err := a.searchClient.Search(ctx, subQuery, 5, true)
			if err != nil {
				log.Printf("Sub-query search failed: %v", err)
				continue
			}
			
			allResults = append(allResults, results.Results...)
		}
		
		reasoningSteps = append(reasoningSteps, fmt.Sprintf("✅ Собрано %d источников из всех подзапросов", len(allResults)))
	} else {
		// Обычный поиск
		reasoningSteps = append(reasoningSteps, fmt.Sprintf("🔎 Ищу информацию по запросу: %s", searchQuery))
		
		searchResults, err := a.searchClient.Search(ctx, searchQuery, 15, true)
		if err != nil {
			return nil, fmt.Errorf("search failed: %w", err)
		}
		
		allResults = searchResults.Results
		reasoningSteps = append(reasoningSteps, fmt.Sprintf("✅ Найдено %d источников", len(allResults)))
	}

	if len(allResults) == 0 {
		return &models.SearchResponse{
			Query:     query,
			Mode:      "pro",
			Answer:    "Не удалось найти релевантную информацию по вашему запросу.",
			Sources:   []models.Source{},
			Reasoning: strings.Join(reasoningSteps, "\n"),
		}, nil
	}

	// Step 3: Semantic Reranking с BM25
	reasoningSteps = append(reasoningSteps, "🎯 Применяю семантическую переоценку результатов (BM25)")
	allResults = a.reranker.Rerank(searchQuery, allResults)

	// Step 4: Credibility Scoring
	reasoningSteps = append(reasoningSteps, "⭐ Оцениваю достоверность источников")
	allResults = a.credibilityScorer.RankSources(allResults)

	// Берем топ-10 после reranking
	if len(allResults) > 10 {
		allResults = allResults[:10]
	}

	// Step 5: Cross-verification (проверка консистентности)
	reasoningSteps = append(reasoningSteps, "🔍 Проверяю консистентность информации между источниками")
	verification := a.crossVerify(allResults)
	if verification != "" {
		reasoningSteps = append(reasoningSteps, verification)
	}

	// Step 6: Format sources for LLM
	var sourcesContext strings.Builder
	for i, result := range allResults {
		if i >= 8 {
			break
		}
		content := result.Content
		if result.RawContent != "" {
			content = result.RawContent
		}
		if len(content) > 800 {
			content = content[:800]
		}
		sourcesContext.WriteString(fmt.Sprintf(
			"Источник %d [Достоверность: %.2f] (%s):\n%s\n\n",
			i+1, result.Credibility, result.Title, content,
		))
	}

	// Step 7: Build LLM prompt
	var promptBuilder strings.Builder
	promptBuilder.WriteString(`Ты исследовательский ассистент в режиме Pro с глубоким анализом.

Твоя задача:
1. Дать подробный, хорошо обоснованный ответ
2. Использовать информацию из источников с учетом их достоверности
3. Указать, если информация противоречива или недостаточна
4. Делать выводы на основе перекрестной проверки

`)

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
	promptBuilder.WriteString("Найденная информация (отсортирована по релевантности и достоверности):\n")
	promptBuilder.WriteString(sourcesContext.String())
	promptBuilder.WriteString("\nПодробный ответ с анализом:")

	reasoningSteps = append(reasoningSteps, "💡 Формирую финальный ответ с учётом всех данных...")

	// Step 8: Generate answer
	answer, err := a.llmClient.Complete(ctx, promptBuilder.String(), 0.7, 1200)
	if err != nil {
		return nil, fmt.Errorf("LLM completion failed: %w", err)
	}

	// Step 9: Format sources
	sources := make([]models.Source, 0)
	for i, result := range allResults {
		if i >= 8 {
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
			Credibility: result.Credibility,
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

// detectMultiHop определяет, нужен ли multi-hop reasoning
func (a *ProAgent) detectMultiHop(query string) bool {
	queryLower := strings.ToLower(query)
	
	multiHopIndicators := []string{
		"сравни", "отличия", "различия", "разница между",
		"как связаны", "взаимосвязь", "влияние",
		"причины и следствия", "что привело к",
		"этапы", "процесс", "как происходит",
		"compare", "difference", "relationship",
		"causes and effects", "process of",
	}
	
	for _, indicator := range multiHopIndicators {
		if strings.Contains(queryLower, indicator) {
			return true
		}
	}
	
	// Если вопрос длинный и содержит несколько смысловых единиц
	words := strings.Fields(query)
	if len(words) > 15 {
		return true
	}
	
	return false
}

// generateSubQueries разбивает сложный вопрос на подвопросы
func (a *ProAgent) generateSubQueries(ctx context.Context, query string) []string {
	prompt := fmt.Sprintf(`Разбей сложный вопрос на 2-3 простых подвопроса для поиска информации.

Вопрос: %s

Подвопросы (каждый с новой строки, без нумерации):`, query)

	response, err := a.llmClient.Complete(ctx, prompt, 0.3, 300)
	if err != nil {
		log.Printf("Failed to generate sub-queries: %v", err)
		return []string{query}
	}

	lines := strings.Split(response, "\n")
	subQueries := make([]string, 0)
	
	for _, line := range lines {
		line = strings.TrimSpace(line)
		// Удаляем нумерацию если есть
		line = strings.TrimPrefix(line, "- ")
		line = strings.TrimPrefix(line, "• ")
		if len(line) > 10 {
			subQueries = append(subQueries, line)
		}
	}

	if len(subQueries) == 0 {
		return []string{query}
	}

	// Ограничиваем до 3 подвопросов
	if len(subQueries) > 3 {
		subQueries = subQueries[:3]
	}

	return subQueries
}

// crossVerify проверяет консистентность между источниками
func (a *ProAgent) crossVerify(results []models.TavilyResult) string {
	if len(results) < 2 {
		return ""
	}

	// Простая эвристика - ищем повторяющиеся факты
	commonPhrases := make(map[string]int)
	
	for _, result := range results {
		words := strings.Fields(strings.ToLower(result.Content))
		
		// Ищем фразы из 3-4 слов
		for i := 0; i < len(words)-2; i++ {
			phrase := strings.Join(words[i:i+3], " ")
			if len(phrase) > 15 { // минимальная длина фразы
				commonPhrases[phrase]++
			}
		}
	}

	// Считаем сколько фактов подтверждены несколькими источниками
	verifiedCount := 0
	for _, count := range commonPhrases {
		if count >= 2 {
			verifiedCount++
		}
	}

	if verifiedCount > 3 {
		return fmt.Sprintf("✓ Найдено %d+ фактов, подтвержденных несколькими источниками", verifiedCount)
	} else if verifiedCount > 0 {
		return "⚠️ Некоторые факты подтверждены только одним источником"
	}

	return "⚠️ Источники содержат разную информацию - требуется дополнительная проверка"
}