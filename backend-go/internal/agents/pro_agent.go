package agents

import (
	"context"
	"fmt"
	"log"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/Vova4o/bootCamp2025CaseSber/backend/internal/models"
	"github.com/Vova4o/bootCamp2025CaseSber/backend/internal/tools"
	"github.com/Vova4o/bootCamp2025CaseSber/backend/internal/utils"
)

type ProAgent struct {
	searchClient      *tools.SearchClient
	llmClient         *tools.LLMClient
	reranker          *tools.BM25Reranker
	credibilityScorer *tools.CredibilityScorer
	timeout           time.Duration
}

func NewProAgent(searchClient *tools.SearchClient, llmClient *tools.LLMClient) *ProAgent {
	return &ProAgent{
		searchClient:      searchClient,
		llmClient:         llmClient,
		reranker:          tools.NewBM25Reranker(),
		credibilityScorer: tools.NewCredibilityScorer(),
		timeout:           20 * time.Second, // Global timeout
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
	// Apply global timeout
	ctx, cancel := context.WithTimeout(ctx, a.timeout)
	defer cancel()

	queryLang := detectLanguage(query)
	log.Printf("Pro mode processing: %s (lang: %s, with context: %v)",
		query, queryLang, len(conversationHistory) > 0)

	reasoningSteps := []string{}
	searchQuery := query

	// Step 1: Enhance query with context
	if len(conversationHistory) > 0 {
		if queryLang == "ru" {
			reasoningSteps = append(reasoningSteps, "🔍 Анализирую контекст предыдущего диалога...")
		} else {
			reasoningSteps = append(reasoningSteps, "🔍 Analyzing previous conversation context...")
		}

		var contextPrompt strings.Builder
		if queryLang == "ru" {
			contextPrompt.WriteString("Предыдущая беседа:\n")
		} else {
			contextPrompt.WriteString("Previous conversation:\n")
		}

		start := len(conversationHistory) - 6
		if start < 0 {
			start = 0
		}
		for _, msg := range conversationHistory[start:] {
			role := msg.Role
			if queryLang == "ru" {
				if msg.Role == "user" {
					role = "Пользователь"
				} else {
					role = "Ассистент"
				}
			}
			contextPrompt.WriteString(fmt.Sprintf("\n%s: %s\n", role, msg.Content))
		}

		var enhancePrompt string
		if queryLang == "ru" {
			enhancePrompt = fmt.Sprintf(`%s

Текущий вопрос: %s

Перефразируй текущий вопрос так, чтобы он был самодостаточным и включал важную информацию из контекста. Улучшенный поисковый запрос:`, contextPrompt.String(), query)
		} else {
			enhancePrompt = fmt.Sprintf(`%s

Current question: %s

Rephrase the current question to be self-contained and include important information from context. Enhanced search query:`, contextPrompt.String(), query)
		}

		enhanced, err := a.llmClient.Complete(ctx, enhancePrompt, 0.3, 200)
		if err != nil {
			log.Printf("⚠️  LLM failed to enhance query, using original: %v", err)
			if queryLang == "ru" {
				reasoningSteps = append(reasoningSteps, "⚠️ Использую оригинальный запрос (LLM недоступен)")
			} else {
				reasoningSteps = append(reasoningSteps, "⚠️ Using original query (LLM unavailable)")
			}
		} else if enhanced != "" {
			searchQuery = strings.TrimSpace(enhanced)
			searchQuery = strings.Trim(searchQuery, `"'`)
			searchQuery = strings.TrimSpace(searchQuery)

			if searchQuery == "" {
				searchQuery = query
				log.Printf("⚠️  Enhanced query was empty after cleanup")
			}

			if queryLang == "ru" {
				reasoningSteps = append(reasoningSteps, fmt.Sprintf("✨ Улучшенный запрос: \"%s\"", searchQuery))
			} else {
				reasoningSteps = append(reasoningSteps, fmt.Sprintf("✨ Enhanced query: \"%s\"", searchQuery))
			}
		} else {
			log.Printf("⚠️  LLM returned empty enhanced query")
			if queryLang == "ru" {
				reasoningSteps = append(reasoningSteps, "⚠️ Использую оригинальный запрос")
			} else {
				reasoningSteps = append(reasoningSteps, "⚠️ Using original query")
			}
		}
	} else {
		if queryLang == "ru" {
			reasoningSteps = append(reasoningSteps, "📝 Обрабатываю первый запрос без контекста")
		} else {
			reasoningSteps = append(reasoningSteps, "📝 Processing first query without context")
		}
	}

	// Step 2: Detect if multi-hop is needed
	needsMultiHop := a.detectMultiHop(query)

	var allResults []models.TavilyResult

	if needsMultiHop {
		if queryLang == "ru" {
			reasoningSteps = append(reasoningSteps, "🔬 Обнаружен сложный вопрос - применяю multi-hop reasoning")
		} else {
			reasoningSteps = append(reasoningSteps, "🔬 Complex question detected - applying multi-hop reasoning")
		}

		subQueries := a.generateSubQueries(ctx, searchQuery, queryLang)
		if queryLang == "ru" {
			reasoningSteps = append(reasoningSteps, fmt.Sprintf("📋 Разбил на %d подвопроса", len(subQueries)))
		} else {
			reasoningSteps = append(reasoningSteps, fmt.Sprintf("📋 Split into %d sub-questions", len(subQueries)))
		}

		// Try parallel search
		allResults = a.parallelSubQuerySearch(ctx, subQueries, queryLang, &reasoningSteps)

		// FALLBACK: If insufficient results from multi-hop
		if len(allResults) < 3 {
			log.Printf("🔄 Multi-hop insufficient results (%d), falling back to direct search", len(allResults))

			if queryLang == "ru" {
				reasoningSteps = append(reasoningSteps,
					fmt.Sprintf("🔄 Недостаточно результатов (%d), выполняю прямой поиск", len(allResults)))
			} else {
				reasoningSteps = append(reasoningSteps,
					fmt.Sprintf("🔄 Insufficient results (%d), performing direct search", len(allResults)))
			}

			directResults, err := a.searchClient.Search(ctx, searchQuery, 15, true)
			if err != nil {
				log.Printf("❌ Fallback search also failed: %v", err)
				// Return what we have from multi-hop
			} else {
				// Merge results, prioritizing multi-hop
				allResults = append(allResults, directResults.Results...)
				log.Printf("✅ Fallback search added %d results", len(directResults.Results))
			}
		}

		if queryLang == "ru" {
			reasoningSteps = append(reasoningSteps,
				fmt.Sprintf("📚 Собрано %d источников", len(allResults)))
		} else {
			reasoningSteps = append(reasoningSteps,
				fmt.Sprintf("📚 Collected %d sources", len(allResults)))
		}
	} else {
		// Regular search
		log.Printf("🔎 Executing search with query: %s", searchQuery)
		if queryLang == "ru" {
			reasoningSteps = append(reasoningSteps, fmt.Sprintf("🔎 Ищу информацию по запросу: \"%s\"", searchQuery))
		} else {
			reasoningSteps = append(reasoningSteps, fmt.Sprintf("🔎 Searching for: \"%s\"", searchQuery))
		}

		searchResults, err := a.searchClient.Search(ctx, searchQuery, 15, true)
		if err != nil {
			log.Printf("❌ Search failed: %v", err)
			return nil, fmt.Errorf("search failed: %w", err)
		}

		allResults = searchResults.Results
		log.Printf("✅ Search returned %d results", len(allResults))
		if queryLang == "ru" {
			reasoningSteps = append(reasoningSteps, fmt.Sprintf("✅ Найдено %d источников", len(allResults)))
		} else {
			reasoningSteps = append(reasoningSteps, fmt.Sprintf("✅ Found %d sources", len(allResults)))
		}
	}

	if len(allResults) == 0 {
		var answer string
		if queryLang == "ru" {
			answer = "Не удалось найти релевантную информацию по вашему запросу."
		} else {
			answer = "Could not find relevant information for your query."
		}

		return &models.SearchResponse{
			Query:     query,
			Mode:      "pro",
			Answer:    answer,
			Sources:   []models.Source{},
			Reasoning: strings.Join(reasoningSteps, "\n"),
		}, nil
	}

	// Step 3: Semantic Reranking с BM25
	if queryLang == "ru" {
		reasoningSteps = append(reasoningSteps, "🎯 Применяю семантическую переоценку результатов (BM25)")
	} else {
		reasoningSteps = append(reasoningSteps, "🎯 Applying semantic re-ranking (BM25)")
	}
	allResults = a.reranker.Rerank(searchQuery, allResults)

	// Step 4: Credibility Scoring
	if queryLang == "ru" {
		reasoningSteps = append(reasoningSteps, "⭐ Оцениваю достоверность источников")
	} else {
		reasoningSteps = append(reasoningSteps, "⭐ Evaluating source credibility")
	}
	allResults = a.credibilityScorer.RankSources(allResults)

	// Step 5: Ensure Domain Diversity
	if queryLang == "ru" {
		reasoningSteps = append(reasoningSteps, "🌐 Обеспечиваю разнообразие источников")
	} else {
		reasoningSteps = append(reasoningSteps, "🌐 Ensuring source diversity")
	}
	topResults := a.selectDiverseSources(allResults, 10)

	// Step 6: Cross-verification
	if queryLang == "ru" {
		reasoningSteps = append(reasoningSteps, "🔍 Проверяю консистентность информации между источниками")
	} else {
		reasoningSteps = append(reasoningSteps, "🔍 Cross-verifying information across sources")
	}
	verification := a.crossVerify(topResults, queryLang)
	if verification != "" {
		reasoningSteps = append(reasoningSteps, verification)
	}

	// Step 7: Format sources for LLM (top 8 for context window)
	var sourcesContext strings.Builder
	displaySources := topResults
	if len(displaySources) > 8 {
		displaySources = displaySources[:8]
	}

	for i, result := range displaySources {
		content := result.Content
		if result.RawContent != "" {
			content = result.RawContent
		}
		
		// Sanitize and truncate safely
		content = utils.SanitizeUTF8(content)
		if len(content) > 800 {
			content = utils.TruncateUTF8WithEllipsis(content, 800)
		}

		if queryLang == "ru" {
			sourcesContext.WriteString(fmt.Sprintf(
				"Источник %d [Достоверность: %.2f] (%s):\n%s\n\n",
				i+1, result.Credibility, result.Title, content,
			))
		} else {
			sourcesContext.WriteString(fmt.Sprintf(
				"Source %d [Credibility: %.2f] (%s):\n%s\n\n",
				i+1, result.Credibility, result.Title, content,
			))
		}
	}

	// Step 8: Build LLM prompt
	var promptBuilder strings.Builder
	if queryLang == "ru" {
		promptBuilder.WriteString(`Ты исследовательский ассистент в режиме Pro с глубоким анализом.

Твоя задача:
1. Дать подробный, хорошо обоснованный ответ
2. Использовать информацию из источников с учетом их достоверности
3. Указать, если информация противоречива или недостаточна
4. Делать выводы на основе перекрестной проверки

`)
	} else {
		promptBuilder.WriteString(`You are a Pro research assistant with deep analysis capabilities.

Your task:
1. Provide a detailed, well-reasoned answer
2. Use information from sources considering their credibility
3. Indicate if information is contradictory or insufficient
4. Draw conclusions based on cross-verification

`)
	}

	if len(conversationHistory) > 0 {
		if queryLang == "ru" {
			promptBuilder.WriteString("\nКонтекст диалога:\n")
		} else {
			promptBuilder.WriteString("\nConversation context:\n")
		}
		start := len(conversationHistory) - 4
		if start < 0 {
			start = 0
		}
		for _, msg := range conversationHistory[start:] {
			promptBuilder.WriteString(fmt.Sprintf("%s: %s\n", msg.Role, msg.Content))
		}
		promptBuilder.WriteString("\n")
	}

	if queryLang == "ru" {
		promptBuilder.WriteString(fmt.Sprintf("Вопрос: %s\n\n", query))
		promptBuilder.WriteString("Найденная информация (отсортирована по релевантности и достоверности):\n")
		promptBuilder.WriteString(sourcesContext.String())
		promptBuilder.WriteString("\nПодробный ответ с анализом:")
	} else {
		promptBuilder.WriteString(fmt.Sprintf("Question: %s\n\n", query))
		promptBuilder.WriteString("Found information (sorted by relevance and credibility):\n")
		promptBuilder.WriteString(sourcesContext.String())
		promptBuilder.WriteString("\nDetailed answer with analysis:")
	}

	if queryLang == "ru" {
		reasoningSteps = append(reasoningSteps, "💡 Формирую финальный ответ с учётом всех данных...")
	} else {
		reasoningSteps = append(reasoningSteps, "💡 Generating final answer based on all data...")
	}

	// Step 9: Generate answer
	answer, err := a.llmClient.Complete(ctx, promptBuilder.String(), 0.7, 1200)
	if err != nil {
		return nil, fmt.Errorf("LLM completion failed: %w", err)
	}

	// Step 10: Format sources with UTF-8 safety
	sources := make([]models.Source, 0)
	for i, result := range displaySources {
		if i >= 8 {
			break
		}
		
		snippet := utils.SanitizeUTF8(result.Snippet)
		if len(snippet) > 200 {
			snippet = utils.TruncateUTF8WithEllipsis(snippet, 200)
		}
		
		sources = append(sources, models.Source{
			Title:       utils.SanitizeUTF8(result.Title),
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

// parallelSubQuerySearch performs parallel searches for sub-queries
func (a *ProAgent) parallelSubQuerySearch(
	ctx context.Context,
	subQueries []string,
	queryLang string,
	reasoningSteps *[]string,
) []models.TavilyResult {
	type searchResult struct {
		results []models.TavilyResult
		query   string
		err     error
	}

	// Try parallel search with extended timeout
	resultsChan := make(chan searchResult, len(subQueries))
	var wg sync.WaitGroup

	for _, subQuery := range subQueries {
		wg.Add(1)
		go func(q string) {
			defer wg.Done()

			// Increased per-query timeout to handle slow responses
			queryCtx, cancel := context.WithTimeout(ctx, 12*time.Second)
			defer cancel()

			res, err := a.searchClient.Search(queryCtx, q, 5, true)
			if err != nil {
				log.Printf("Sub-query search failed for '%s': %v", q, err)
				resultsChan <- searchResult{nil, q, err}
				return
			}

			resultsChan <- searchResult{res.Results, q, nil}
		}(subQuery)
	}

	// Close channel when all goroutines finish
	go func() {
		wg.Wait()
		close(resultsChan)
	}()

	// Collect results
	allResults := make([]models.TavilyResult, 0)
	successCount := 0
	failCount := 0

	for sr := range resultsChan {
		if sr.err != nil {
			failCount++
			if queryLang == "ru" {
				*reasoningSteps = append(*reasoningSteps,
					fmt.Sprintf("  ⚠️ Подзапрос пропущен (timeout): %s",
						truncateQuery(sr.query, 60)))
			} else {
				*reasoningSteps = append(*reasoningSteps,
					fmt.Sprintf("  ⚠️ Sub-query skipped (timeout): %s",
						truncateQuery(sr.query, 60)))
			}
			continue
		}

		successCount++
		if queryLang == "ru" {
			*reasoningSteps = append(*reasoningSteps,
				fmt.Sprintf("  ✓ %s (%d результатов)",
					truncateQuery(sr.query, 60), len(sr.results)))
		} else {
			*reasoningSteps = append(*reasoningSteps,
				fmt.Sprintf("  ✓ %s (%d results)",
					truncateQuery(sr.query, 60), len(sr.results)))
		}

		allResults = append(allResults, sr.results...)
	}

	// FALLBACK: If most sub-queries failed or not enough results
	if failCount >= len(subQueries)/2 || len(allResults) < 3 {
		log.Printf("⚠️ Multi-hop fallback: %d/%d sub-queries failed, switching to direct search",
			failCount, len(subQueries))

		if queryLang == "ru" {
			*reasoningSteps = append(*reasoningSteps,
				fmt.Sprintf("⚠️ Переключаюсь на прямой поиск (подзапросы: успех %d, фейл %d)",
					successCount, failCount))
		} else {
			*reasoningSteps = append(*reasoningSteps,
				fmt.Sprintf("⚠️ Switching to direct search (sub-queries: success %d, failed %d)",
					successCount, failCount))
		}

		return allResults // Return partial results, caller will handle direct search
	}

	return allResults
}

// Helper function to truncate long queries
func truncateQuery(query string, maxLen int) string {
	return utils.TruncateUTF8(query, maxLen)
}

// selectDiverseSources ensures domain diversity in results
func (a *ProAgent) selectDiverseSources(results []models.TavilyResult, maxResults int) []models.TavilyResult {
	selected := make([]models.TavilyResult, 0, maxResults)
	domainCounts := make(map[string]int)
	maxPerDomain := 2 // Maximum 2 results from same domain

	for _, result := range results {
		if len(selected) >= maxResults {
			break
		}

		domain := extractDomain(result.URL)
		if domain == "" {
			continue
		}

		// Allow up to maxPerDomain results from same domain
		if domainCounts[domain] < maxPerDomain {
			selected = append(selected, result)
			domainCounts[domain]++
		}
	}

	// If we didn't get enough diverse results, fill remaining slots
	if len(selected) < maxResults {
		for _, result := range results {
			if len(selected) >= maxResults {
				break
			}

			domain := extractDomain(result.URL)
			if domainCounts[domain] >= maxPerDomain {
				// Allow one more from this domain
				found := false
				for _, s := range selected {
					if s.URL == result.URL {
						found = true
						break
					}
				}
				if !found {
					selected = append(selected, result)
				}
			}
		}
	}

	log.Printf("📊 Domain diversity: %d unique domains from %d sources",
		len(domainCounts), len(selected))

	return selected
}

// detectMultiHop determines if multi-hop reasoning is needed (improved)
func (a *ProAgent) detectMultiHop(query string) bool {
	queryLower := strings.ToLower(query)

	// Strong indicators for multi-hop
	strongIndicators := []string{
		// Comparison
		"сравни", "compare", "отличия", "difference", "различия",
		"разница между", "difference between",
		// Causation
		"как связаны", "relationship", "взаимосвязь",
		"влияние", "influence", "impact",
		"причины и следствия", "causes and effects",
		"что привело к", "what led to", "how did",
		// Analysis
		"advantages and disadvantages", "pros and cons",
		"преимущества и недостатки", "за и против",
	}

	for _, indicator := range strongIndicators {
		if strings.Contains(queryLower, indicator) {
			return true
		}
	}

	// Only for VERY long queries with multiple concepts
	words := strings.Fields(query)
	if len(words) > 20 {
		// Check if query has multiple question words/concepts
		questionWords := 0
		for _, word := range []string{"what", "how", "why", "когда", "как", "почему", "что"} {
			if strings.Contains(queryLower, word) {
				questionWords++
			}
		}
		return questionWords >= 2
	}

	return false
}

// generateSubQueries splits complex query into sub-questions
func (a *ProAgent) generateSubQueries(ctx context.Context, query string, lang string) []string {
	var prompt string
	if lang == "ru" {
		prompt = fmt.Sprintf(`Разбей сложный вопрос на 2-3 простых подвопроса для поиска информации.

Вопрос: %s

Подвопросы (каждый с новой строки, без нумерации):`, query)
	} else {
		prompt = fmt.Sprintf(`Break down this complex question into 2-3 simple sub-questions for information search.

Question: %s

Sub-questions (one per line, no numbering):`, query)
	}

	response, err := a.llmClient.Complete(ctx, prompt, 0.3, 300)
	if err != nil {
		log.Printf("Failed to generate sub-queries: %v", err)
		return []string{query}
	}

	lines := strings.Split(response, "\n")
	subQueries := make([]string, 0)

	for _, line := range lines {
		line = strings.TrimSpace(line)
		// Remove various prefixes
		line = strings.TrimPrefix(line, "- ")
		line = strings.TrimPrefix(line, "• ")
		line = strings.TrimPrefix(line, "* ")

		// Remove numbering
		for i := 1; i <= 9; i++ {
			line = strings.TrimPrefix(line, fmt.Sprintf("%d. ", i))
			line = strings.TrimPrefix(line, fmt.Sprintf("%d) ", i))
		}

		line = strings.TrimSpace(line)

		// Only add substantial queries
		if len(line) > 10 &&
			!strings.Contains(strings.ToLower(line), "sub-question") &&
			!strings.Contains(strings.ToLower(line), "подвопрос") {
			subQueries = append(subQueries, line)
		}
	}

	if len(subQueries) == 0 {
		return []string{query}
	}

	// Limit to 3 sub-queries
	if len(subQueries) > 3 {
		subQueries = subQueries[:3]
	}

	return subQueries
}

// crossVerify checks consistency between sources
func (a *ProAgent) crossVerify(results []models.TavilyResult, lang string) string {
	if len(results) < 2 {
		return ""
	}

	commonPhrases := make(map[string]int)

	for _, result := range results {
		words := strings.Fields(strings.ToLower(result.Content))

		// Look for 3-4 word phrases
		for i := 0; i < len(words)-2; i++ {
			phrase := strings.Join(words[i:i+3], " ")
			if len(phrase) > 15 {
				commonPhrases[phrase]++
			}
		}
	}

	// Count facts verified by multiple sources
	verifiedCount := 0
	for _, count := range commonPhrases {
		if count >= 2 {
			verifiedCount++
		}
	}

	if lang == "ru" {
		if verifiedCount > 3 {
			return fmt.Sprintf("✓ Найдено %d+ фактов, подтвержденных несколькими источниками", verifiedCount)
		} else if verifiedCount > 0 {
			return "⚠️ Некоторые факты подтверждены только одним источником"
		}
		return "⚠️ Источники содержат разную информацию - требуется дополнительная проверка"
	} else {
		if verifiedCount > 3 {
			return fmt.Sprintf("✓ Found %d+ facts verified by multiple sources", verifiedCount)
		} else if verifiedCount > 0 {
			return "⚠️ Some facts verified by only one source"
		}
		return "⚠️ Sources contain different information - additional verification needed"
	}
}

// detectLanguage determines text language
func detectLanguage(text string) string {
	cyrillicCount := 0
	totalLetters := 0

	for _, r := range text {
		if (r >= 'а' && r <= 'я') || (r >= 'А' && r <= 'Я') || r == 'ё' || r == 'Ё' {
			cyrillicCount++
			totalLetters++
		} else if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') {
			totalLetters++
		}
	}

	if totalLetters == 0 {
		return "en"
	}

	if float64(cyrillicCount)/float64(totalLetters) > 0.3 {
		return "ru"
	}

	return "en"
}

// extractDomain extracts clean domain from URL
func extractDomain(urlStr string) string {
	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		return ""
	}

	hostname := strings.ToLower(parsedURL.Hostname())

	// Remove www. prefix
	hostname = strings.TrimPrefix(hostname, "www.")

	return hostname
}