package agents

import (
	"context"
	"log"
	"strings"

	"github.com/Vova4o/bootCamp2025CaseSber/backend/internal/tools"
)

type ModeSelector struct {
	llmClient *tools.LLMClient
}

func NewModeSelector(llmClient *tools.LLMClient) *ModeSelector {
	return &ModeSelector{llmClient: llmClient}
}

func (m *ModeSelector) SelectMode(ctx context.Context, query string) (string, error) {
	queryLower := strings.ToLower(query)

	// Simple heuristics for quick classification
	simpleIndicators := []string{
		"кто", "что такое", "когда", "где", "сколько",
		"какой", "какая", "какое", "как зовут",
		"столица", "год", "дата", "возраст",
		"погода", "курс", "цена",
		"who", "what is", "when", "where", "how much",
		"capital", "weather", "price",
	}

	complexIndicators := []string{
		"сравни", "проанализируй", "объясни почему",
		"различия между", "преимущества и недостатки",
		"как работает", "причины", "последствия",
		"влияние", "взаимосвязь", "теории",
		"compare", "analyze", "explain why",
		"differences between", "advantages and disadvantages",
		"how does", "causes", "consequences",
	}

	hasSimple := containsAny(queryLower, simpleIndicators)
	hasComplex := containsAny(queryLower, complexIndicators)

	// Quick decision for obvious cases
	if hasSimple && !hasComplex && len(strings.Split(query, " ")) < 10 {
		log.Printf("Query classified as SIMPLE (heuristic): %s", query)
		return "simple", nil
	}

	if hasComplex {
		log.Printf("Query classified as PRO (heuristic): %s", query)
		return "pro", nil
	}

	// Use LLM for borderline cases
	prompt := `Ты классификатор запросов. Определи сложность запроса.

SIMPLE - для простых фактических вопросов:
- Кто президент США?
- Когда основан Google?
- Столица Франции?
- Погода в Москве?

PRO - для сложных аналитических вопросов:
- Сравни подходы к регулированию AI
- Объясни причины экономического кризиса 2008
- Проанализируй влияние социальных сетей на общество

Запрос: ` + query + `

Ответь ТОЛЬКО одним словом: SIMPLE или PRO`

	response, err := m.llmClient.Complete(ctx, prompt, 0.1, 10)
	if err != nil {
		log.Printf("LLM mode selection failed: %v, defaulting to simple", err)
		return "simple", nil
	}

	mode := "simple"
	if strings.Contains(strings.ToUpper(response), "PRO") {
		mode = "pro"
	}

	log.Printf("Query classified as %s (LLM): %s", strings.ToUpper(mode), query)
	return mode, nil
}

func containsAny(text string, indicators []string) bool {
	for _, indicator := range indicators {
		if strings.Contains(text, indicator) {
			return true
		}
	}
	return false
}
