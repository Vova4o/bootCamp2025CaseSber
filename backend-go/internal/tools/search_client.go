package tools

import (
	"context"
	"fmt"
	"log"
	"net/url"
	"regexp"
	"strings"

	"github.com/Vova4o/bootCamp2025CaseSber/backend/internal/models"
	"github.com/go-resty/resty/v2"
)

type SearchClient struct {
	client *resty.Client
}

func NewSearchClient() *SearchClient {
	client := resty.New()
	client.SetHeader("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36")
	return &SearchClient{
		client: client,
	}
}

// DuckDuckGo VQD token response
type DDGVQDResponse struct {
	VQD string `json:"vqd"`
}

func (s *SearchClient) Search(
	ctx context.Context,
	query string,
	maxResults int,
	includeRawContent bool,
) (*models.TavilySearchResponse, error) {
	log.Printf("🔍 Searching DuckDuckGo for: %s", query)

	// Try Instant Answer API first
	instantResults := s.tryInstantAnswer(ctx, query, maxResults)
	if len(instantResults) > 0 {
		log.Printf("✅ Found %d results from Instant Answer API", len(instantResults))
		return &models.TavilySearchResponse{Results: instantResults}, nil
	}

	// Fallback to HTML search
	htmlResults := s.tryHTMLSearch(ctx, query, maxResults)
	if len(htmlResults) > 0 {
		log.Printf("✅ Found %d results from HTML search", len(htmlResults))
		return &models.TavilySearchResponse{Results: htmlResults}, nil
	}

	log.Printf("⚠️ No results found for query: %s", query)
	return &models.TavilySearchResponse{Results: []models.TavilyResult{}}, nil
}

func (s *SearchClient) tryInstantAnswer(ctx context.Context, query string, maxResults int) []models.TavilyResult {
	type DDGResponse struct {
		RelatedTopics []struct {
			FirstURL string `json:"FirstURL"`
			Text     string `json:"Text"`
		} `json:"RelatedTopics"`
		Abstract     string `json:"Abstract"`
		AbstractURL  string `json:"AbstractURL"`
		AbstractText string `json:"AbstractText"`
	}

	ddgURL := fmt.Sprintf("https://api.duckduckgo.com/?q=%s&format=json&no_html=1&skip_disambig=1",
		url.QueryEscape(query))

	var ddgResp DDGResponse
	resp, err := s.client.R().
		SetContext(ctx).
		SetResult(&ddgResp).
		Get(ddgURL)

	if err != nil || resp.IsError() {
		return nil
	}

	results := make([]models.TavilyResult, 0, maxResults)

	// Add abstract as first result
	if ddgResp.AbstractText != "" && ddgResp.AbstractURL != "" {
		results = append(results, models.TavilyResult{
			Title:   "Основная информация",
			URL:     ddgResp.AbstractURL,
			Content: ddgResp.AbstractText,
			Score:   1.0,
		})
	}

	// Add related topics
	for i, topic := range ddgResp.RelatedTopics {
		if len(results) >= maxResults {
			break
		}
		if topic.FirstURL != "" && topic.Text != "" {
			title := topic.Text
			if len(title) > 100 {
				title = title[:97] + "..."
			}
			results = append(results, models.TavilyResult{
				Title:   title,
				URL:     topic.FirstURL,
				Content: topic.Text,
				Score:   0.9 - float64(i)*0.1,
			})
		}
	}

	return results
}

func (s *SearchClient) tryHTMLSearch(ctx context.Context, query string, maxResults int) []models.TavilyResult {
	// Get HTML search results
	searchURL := fmt.Sprintf("https://html.duckduckgo.com/html/?q=%s", url.QueryEscape(query))

	resp, err := s.client.R().
		SetContext(ctx).
		SetHeader("Accept", "text/html").
		Get(searchURL)

	if err != nil || resp.IsError() {
		log.Printf("HTML search failed: %v", err)
		return nil
	}

	html := resp.String()
	results := make([]models.TavilyResult, 0, maxResults)

	// Parse HTML results using regex (simple parsing)
	resultPattern := regexp.MustCompile(`<a rel="nofollow" class="result__a" href="([^"]+)">([^<]+)</a>`)
	snippetPattern := regexp.MustCompile(`<a class="result__snippet"[^>]*>([^<]+)</a>`)

	matches := resultPattern.FindAllStringSubmatch(html, -1)
	snippets := snippetPattern.FindAllStringSubmatch(html, -1)

	for i := 0; i < len(matches) && i < maxResults; i++ {
		if len(matches[i]) < 3 {
			continue
		}

		resultURL := strings.TrimSpace(matches[i][1])
		title := strings.TrimSpace(matches[i][2])

		// Extract snippet
		snippet := ""
		if i < len(snippets) && len(snippets[i]) > 1 {
			snippet = strings.TrimSpace(snippets[i][1])
		}

		// Clean up URL (DuckDuckGo redirects)
		if strings.HasPrefix(resultURL, "//duckduckgo.com/l/?") {
			if u, err := url.Parse("https:" + resultURL); err == nil {
				if uddParam := u.Query().Get("uddg"); uddParam != "" {
					if decoded, err := url.QueryUnescape(uddParam); err == nil {
						resultURL = decoded
					}
				}
			}
		}

		results = append(results, models.TavilyResult{
			Title:   title,
			URL:     resultURL,
			Content: snippet,
			Score:   1.0 - float64(i)*0.1,
		})
	}

	return results
}
