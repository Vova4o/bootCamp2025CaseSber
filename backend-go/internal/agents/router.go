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
	cfg          *config.Config
	searchClient *tools.SearchClient
	llmClient    *tools.LLMClient
	simpleAgent  *SimpleAgent
	proAgent     *ProAgent
	modeSelector *ModeSelector
}

func NewRouterAgent(cfg *config.Config) *RouterAgent {
	searchClient := tools.NewSearchClient()
	llmClient := tools.NewLLMClient(cfg)

	return &RouterAgent{
		cfg:          cfg,
		searchClient: searchClient,
		llmClient:    llmClient,
		simpleAgent:  NewSimpleAgent(searchClient, llmClient),
		proAgent:     NewProAgent(searchClient, llmClient),
		modeSelector: NewModeSelector(llmClient),
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
		var err error
		selectedMode, err = r.modeSelector.SelectMode(ctx, query)
		if err != nil {
			log.Printf("Mode selection failed, defaulting to simple: %v", err)
			selectedMode = "simple"
		}
		log.Printf("Auto mode selected: %s for query: %s", selectedMode, query)
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
	case "simple":
		result, err = r.simpleAgent.Process(ctx, query)
	default:
		return nil, fmt.Errorf("unknown mode: %s", selectedMode)
	}

	if err != nil {
		return nil, err
	}

	result.Mode = selectedMode
	return result, nil
}
