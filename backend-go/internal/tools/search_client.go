package tools

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/Vova4o/bootCamp2025CaseSber/backend/internal/models"
	"github.com/go-resty/resty/v2"
)

type SearchClient struct {
	client      *resty.Client
	userAgents  []string
	lastReqTime time.Time
	searxngURL  string
	braveAPIKey string
}

func NewSearchClient() *SearchClient {
	client := resty.New()
	client.SetTimeout(20 * time.Second)
	client.SetRetryCount(3)
	client.SetRetryWaitTime(2 * time.Second)

	searxngURL := os.Getenv("SEARXNG_URL")
	if searxngURL == "" {
		searxngURL = "http://searxng:8080" // Docker service name
	}

	return &SearchClient{
		client:      client,
		searxngURL:  searxngURL,
		braveAPIKey: os.Getenv("BRAVE_SEARCH_API_KEY"),
		userAgents: []string{
			"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36",
			"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36",
			"Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36",
		},
	}
}

func (s *SearchClient) getRandomUserAgent() string {
	return s.userAgents[rand.Intn(len(s.userAgents))]
}

func (s *SearchClient) rateLimit() {
	elapsed := time.Since(s.lastReqTime)
	minDelay := 500 * time.Millisecond

	if elapsed < minDelay {
		time.Sleep(minDelay - elapsed)
	}
	s.lastReqTime = time.Now()
}

func (s *SearchClient) Search(
	ctx context.Context,
	query string,
	maxResults int,
	includeRawContent bool,
) (*models.TavilySearchResponse, error) {
	log.Printf("🔍 Multi-source search for: %s", query)

	var allResults []models.TavilyResult

	// Strategy 1: SearXNG (Primary - aggregates multiple search engines)
	searxngResults := s.trySearXNG(ctx, query, maxResults)
	allResults = append(allResults, searxngResults...)
	log.Printf("  📊 SearXNG: %d results", len(searxngResults))

	// Strategy 2: Brave Search API (Fallback)
	if len(allResults) < 3 && s.braveAPIKey != "" {
		s.rateLimit()
		braveResults := s.tryBraveSearchAPI(ctx, query, maxResults-len(allResults))
		allResults = append(allResults, braveResults...)
		log.Printf("  📊 Brave API: %d results", len(braveResults))
	}

	// Strategy 3: DuckDuckGo Instant Answer (Additional fallback)
	if len(allResults) < 2 {
		s.rateLimit()
		instantResults := s.tryInstantAnswer(ctx, query, maxResults-len(allResults))
		allResults = append(allResults, instantResults...)
		log.Printf("  📊 DDG Instant: %d results", len(instantResults))
	}

	// Strategy 4: DuckDuckGo HTML (Last resort)
	if len(allResults) < 1 {
		s.rateLimit()
		htmlResults := s.tryDDGHTML(ctx, query, maxResults-len(allResults))
		allResults = append(allResults, htmlResults...)
		log.Printf("  📊 DDG HTML: %d results", len(htmlResults))
	}

	// Deduplicate and limit
	allResults = s.deduplicateResults(allResults)

	if len(allResults) > maxResults {
		allResults = allResults[:maxResults]
	}

	log.Printf("✅ Total: %d unique results", len(allResults))
	return &models.TavilySearchResponse{
		Results: allResults,
		Query:   query,
	}, nil
}

// SearXNG search (Primary method)
func (s *SearchClient) trySearXNG(
	ctx context.Context,
	query string,
	maxResults int,
) []models.TavilyResult {
	type SearXNGResponse struct {
		Results []struct {
			Title   string  `json:"title"`
			URL     string  `json:"url"`
			Content string  `json:"content"`
			Engine  string  `json:"engine"`
			Score   float64 `json:"score"`
		} `json:"results"`
		Query string `json:"query"`
	}

	var searxResp SearXNGResponse
	resp, err := s.client.R().
		SetContext(ctx).
		SetQueryParams(map[string]string{
			"q":        query,
			"format":   "json",
			"language": "en",
		}).
		SetResult(&searxResp).
		SetHeader("User-Agent", s.getRandomUserAgent()).
		Get(s.searxngURL + "/search")

	if err != nil {
		log.Printf("⚠️  SearXNG failed: %v", err)
		return nil
	}

	if resp.IsError() {
		log.Printf("⚠️  SearXNG error response: %d", resp.StatusCode())
		return nil
	}

	results := make([]models.TavilyResult, 0)
	for i, r := range searxResp.Results {
		if i >= maxResults {
			break
		}

		// Filter out empty results
		if r.Title == "" || r.URL == "" {
			continue
		}

		content := r.Content
		if content == "" {
			content = r.Title
		}

		// Truncate long content
		if len(content) > 500 {
			content = content[:500] + "..."
		}

		score := 0.95 - float64(i)*0.03
		if r.Score > 0 {
			score = r.Score
		}

		results = append(results, models.TavilyResult{
			Title:   r.Title,
			URL:     r.URL,
			Content: content,
			Snippet: content,
			Score:   score,
		})
	}

	return results
}

// Brave Search API (Fallback)
func (s *SearchClient) tryBraveSearchAPI(
	ctx context.Context,
	query string,
	maxResults int,
) []models.TavilyResult {
	if s.braveAPIKey == "" {
		return nil
	}

	type BraveResponse struct {
		Web struct {
			Results []struct {
				Title       string `json:"title"`
				URL         string `json:"url"`
				Description string `json:"description"`
				Age         string `json:"age"`
			} `json:"results"`
		} `json:"web"`
	}

	var braveResp BraveResponse
	resp, err := s.client.R().
		SetContext(ctx).
		SetHeader("Accept", "application/json").
		SetHeader("Accept-Encoding", "gzip").
		SetHeader("X-Subscription-Token", s.braveAPIKey).
		SetQueryParams(map[string]string{
			"q":     query,
			"count": fmt.Sprintf("%d", maxResults),
		}).
		SetResult(&braveResp).
		Get("https://api.search.brave.com/res/v1/web/search")

	if err != nil {
		log.Printf("⚠️  Brave API failed: %v", err)
		return nil
	}

	if resp.IsError() {
		log.Printf("⚠️  Brave API error: %d - %s", resp.StatusCode(), resp.String())
		return nil
	}

	results := make([]models.TavilyResult, 0)
	for i, r := range braveResp.Web.Results {
		if i >= maxResults {
			break
		}

		if r.Title == "" || r.URL == "" {
			continue
		}

		content := r.Description
		if len(content) > 500 {
			content = content[:500] + "..."
		}

		results = append(results, models.TavilyResult{
			Title:   r.Title,
			URL:     r.URL,
			Content: content,
			Snippet: content,
			Score:   0.9 - float64(i)*0.04,
		})
	}

	return results
}

// DuckDuckGo Instant Answer (Additional fallback)
func (s *SearchClient) tryInstantAnswer(
	ctx context.Context,
	query string,
	maxResults int,
) []models.TavilyResult {
	type DDGResponse struct {
		RelatedTopics []struct {
			FirstURL string `json:"FirstURL"`
			Text     string `json:"Text"`
		} `json:"RelatedTopics"`
		AbstractText string `json:"AbstractText"`
		AbstractURL  string `json:"AbstractURL"`
		Answer       string `json:"Answer"`
	}

	ddgURL := fmt.Sprintf(
		"https://api.duckduckgo.com/?q=%s&format=json&no_html=1&skip_disambig=1",
		url.QueryEscape(query),
	)

	var ddgResp DDGResponse
	resp, err := s.client.R().
		SetContext(ctx).
		SetResult(&ddgResp).
		Get(ddgURL)

	if err != nil || resp.IsError() {
		return nil
	}

	results := make([]models.TavilyResult, 0)

	// Direct answer
	if ddgResp.Answer != "" {
		results = append(results, models.TavilyResult{
			Title:   "Direct Answer",
			URL:     "https://duckduckgo.com",
			Content: ddgResp.Answer,
			Snippet: ddgResp.Answer,
			Score:   1.0,
		})
	}

	// Abstract
	if ddgResp.AbstractText != "" && ddgResp.AbstractURL != "" {
		results = append(results, models.TavilyResult{
			Title:   "Overview",
			URL:     ddgResp.AbstractURL,
			Content: ddgResp.AbstractText,
			Snippet: ddgResp.AbstractText,
			Score:   0.95,
		})
	}

	// Related topics
	for i, topic := range ddgResp.RelatedTopics {
		if len(results) >= maxResults {
			break
		}
		if topic.FirstURL != "" && topic.Text != "" {
			results = append(results, models.TavilyResult{
				Title:   truncateText(topic.Text, 100),
				URL:     topic.FirstURL,
				Content: topic.Text,
				Snippet: topic.Text,
				Score:   0.9 - float64(i)*0.05,
			})
		}
	}

	return results
}

// DuckDuckGo HTML (Last resort)
func (s *SearchClient) tryDDGHTML(
	ctx context.Context,
	query string,
	maxResults int,
) []models.TavilyResult {
	searchURL := fmt.Sprintf(
		"https://html.duckduckgo.com/html/?q=%s",
		url.QueryEscape(query),
	)

	resp, err := s.client.R().
		SetContext(ctx).
		SetHeader("User-Agent", s.getRandomUserAgent()).
		SetHeader("Accept", "text/html,application/xhtml+xml").
		SetHeader("Accept-Language", "en-US,en;q=0.9").
		SetHeader("Referer", "https://duckduckgo.com/").
		Get(searchURL)

	if err != nil || resp.IsError() {
		return nil
	}

	doc, err := goquery.NewDocumentFromReader(
		strings.NewReader(resp.String()),
	)
	if err != nil {
		return nil
	}

	results := make([]models.TavilyResult, 0)

	doc.Find(".result").Each(func(i int, result *goquery.Selection) {
		if len(results) >= maxResults {
			return
		}

		titleLink := result.Find(".result__a")
		title := strings.TrimSpace(titleLink.Text())
		href, _ := titleLink.Attr("href")
		snippet := strings.TrimSpace(result.Find(".result__snippet").Text())

		if title != "" && href != "" {
			if strings.Contains(href, "uddg=") {
				if u, err := url.Parse(href); err == nil {
					if uddg := u.Query().Get("uddg"); uddg != "" {
						if decoded, err := url.QueryUnescape(uddg); err == nil {
							href = decoded
						}
					}
				}
			}

			results = append(results, models.TavilyResult{
				Title:   title,
				URL:     href,
				Content: snippet,
				Snippet: snippet,
				Score:   0.8 - float64(len(results))*0.05,
			})
		}
	})

	return results
}

func (s *SearchClient) deduplicateResults(
	results []models.TavilyResult,
) []models.TavilyResult {
	seen := make(map[string]bool)
	unique := make([]models.TavilyResult, 0)

	for _, result := range results {
		normalizedURL := strings.ToLower(
			strings.TrimRight(result.URL, "/"),
		)

		if !seen[normalizedURL] {
			seen[normalizedURL] = true
			unique = append(unique, result)
		}
	}

	return unique
}

func truncateText(text string, maxLen int) string {
	if len(text) <= maxLen {
		return text
	}
	return text[:maxLen] + "..."
}
