package tools

import (
	"context"
	"fmt"
	"net/url"

	"github.com/Vova4o/bootCamp2025CaseSber/backend/internal/models"
	"github.com/go-resty/resty/v2"
)

type SearchClient struct {
	client *resty.Client
}

func NewSearchClient() *SearchClient {
	return &SearchClient{
		client: resty.New(),
	}
}

// DuckDuckGo Instant Answer API response
type DDGResponse struct {
	RelatedTopics []struct {
		FirstURL string `json:"FirstURL"`
		Text     string `json:"Text"`
	} `json:"RelatedTopics"`
	Abstract     string `json:"Abstract"`
	AbstractURL  string `json:"AbstractURL"`
	AbstractText string `json:"AbstractText"`
}

func (s *SearchClient) Search(
	ctx context.Context,
	query string,
	maxResults int,
	includeRawContent bool,
) (*models.TavilySearchResponse, error) {
	// Use DuckDuckGo Instant Answer API
	ddgURL := fmt.Sprintf("https://api.duckduckgo.com/?q=%s&format=json&no_html=1&skip_disambig=1",
		url.QueryEscape(query))

	var ddgResp DDGResponse
	resp, err := s.client.R().
		SetContext(ctx).
		SetResult(&ddgResp).
		Get(ddgURL)
	if err != nil {
		return nil, fmt.Errorf("DuckDuckGo search failed: %w", err)
	}

	if resp.IsError() {
		return nil, fmt.Errorf("DuckDuckGo API error: status %d", resp.StatusCode())
	}

	// Convert DDG response to our format
	results := make([]models.TavilyResult, 0, maxResults)

	// Add abstract as first result if available
	if ddgResp.AbstractText != "" && ddgResp.AbstractURL != "" {
		results = append(results, models.TavilyResult{
			Title:   "DuckDuckGo Abstract",
			URL:     ddgResp.AbstractURL,
			Content: ddgResp.AbstractText,
			Score:   1.0,
		})
	}

	// Add related topics
	for i, topic := range ddgResp.RelatedTopics {
		if i >= maxResults-1 {
			break
		}
		if topic.FirstURL != "" && topic.Text != "" {
			results = append(results, models.TavilyResult{
				Title:   topic.Text[:min(len(topic.Text), 100)],
				URL:     topic.FirstURL,
				Content: topic.Text,
				Score:   0.8 - float64(i)*0.1,
			})
		}
	}

	return &models.TavilySearchResponse{
		Results: results,
	}, nil
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
