package agents

import (
	"context"
	"fmt"
	"log"

	"github.com/Vova4o/bootCamp2025CaseSber/backend/internal/config"
	"github.com/Vova4o/bootCamp2025CaseSber/backend/internal/models"
	"github.com/Vova4o/bootCamp2025CaseSber/backend/internal/tools"
)

type RouterAgent struct {
	cfg            *config.Config
	searchClient   *tools.SearchClient
	llmClient      *tools.LLMClient
	simpleAgent    *SimpleAgent
	proAgent       *ProAgent
	socialAgent    *SocialAgent
	academicAgent  *AcademicAgent
	financeAgent   *FinanceAgent
	modeSelector   *ModeSelector
}

func NewRouterAgent(cfg *config.Config) *RouterAgent {
	searchClient := tools.NewSearchClient()
	llmClient := tools.NewLLMClient(cfg)

	return &RouterAgent{
		cfg:           cfg,
		searchClient:  searchClient,
		llmClient:     llmClient,
		simpleAgent:   NewSimpleAgent(searchClient, llmClient),
		proAgent:      NewProAgent(searchClient, llmClient),
		socialAgent:   NewSocialAgent(llmClient),
		academicAgent: NewAcademicAgent(llmClient),
		financeAgent:  NewFinanceAgent(llmClient),
		modeSelector:  NewModeSelector(llmClient),
	}
}

func (r *RouterAgent) ProcessQuery(ctx context.Context, query, mode string) (*models.SearchResponse, error) {
	return r.ProcessQueryWithContext(ctx, query, mode, nil)
}

func (r *RouterAgent) ProcessQueryWithContext(
	ctx context.Context,
	query, mode string,
	conversationHistory []models.Message,
) (*models.SearchResponse, error) {
	// Select mode if auto
	selectedMode := mode
	
	if mode == "auto" || mode == "" {
		// AUTO MODE LOGIC: Switch to Pro if context exists
		if len(conversationHistory) > 2 {
			// После 2+ сообщений всегда используем Pro
			selectedMode = "pro"
			log.Printf("🔄 Auto mode: Switching to PRO (context size: %d messages)", len(conversationHistory))
		} else {
			// Первые запросы - выбираем простой/сложный
			var err error
			selectedMode, err = r.modeSelector.SelectMode(ctx, query)
			if err != nil {
				log.Printf("Mode selection failed, defaulting to simple: %v", err)
				selectedMode = "simple"
			}
			log.Printf("🤖 Auto mode selected: %s for query: %s", selectedMode, query)
		}
	}

	// Process based on selected mode
	var result *models.SearchResponse
	var err error

	switch selectedMode {
	case "pro":
		if len(conversationHistory) > 0 {
			result, err = r.proAgent.ProcessWithContext(ctx, query, conversationHistory)
		} else {
			result, err = r.proAgent.Process(ctx, query)
		}
		
	case "pro-social":
		if len(conversationHistory) > 0 {
			result, err = r.socialAgent.ProcessWithContext(ctx, query, conversationHistory)
		} else {
			result, err = r.socialAgent.Process(ctx, query)
		}
		
	case "pro-academic":
		if len(conversationHistory) > 0 {
			result, err = r.academicAgent.ProcessWithContext(ctx, query, conversationHistory)
		} else {
			result, err = r.academicAgent.Process(ctx, query)
		}
		
	case "pro-finance":
		if len(conversationHistory) > 0 {
			result, err = r.financeAgent.ProcessWithContext(ctx, query, conversationHistory)
		} else {
			result, err = r.financeAgent.Process(ctx, query)
		}
		
	case "simple":
		if len(conversationHistory) > 0 {
			result, err = r.simpleAgent.ProcessWithContext(ctx, query, conversationHistory)
		} else {
			result, err = r.simpleAgent.Process(ctx, query)
		}
		
	default:
		return nil, fmt.Errorf("unknown mode: %s", selectedMode)
	}

	if err != nil {
		return nil, err
	}

	// Preserve original mode if it was auto
	if mode == "auto" || mode == "" {
		result.Mode = "auto → " + selectedMode
	} else {
		result.Mode = selectedMode
	}
	
	return result, nil
}