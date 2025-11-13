package scrapers

import (
	"context"
	"fmt"
	"log"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/Vova4o/bootCamp2025CaseSber/backend/internal/models"
	"github.com/go-resty/resty/v2"
)

type SocialScraper struct {
	client *resty.Client
}

func NewSocialScraper() *SocialScraper {
	client := resty.New()
	client.SetTimeout(15 * time.Second)
	client.SetHeader("User-Agent", "Mozilla/5.0 (compatible; ResearchBot/1.0)")
	return &SocialScraper{client: client}
}

// Reddit scraping (без API)
func (s *SocialScraper) SearchReddit(ctx context.Context, query string, limit int) ([]models.TavilyResult, error) {
	log.Printf("🔍 Scraping Reddit for: %s", query)
	
	// Use old.reddit.com for easier parsing
	searchURL := fmt.Sprintf("https://old.reddit.com/search?q=%s&sort=relevance&t=all", 
		url.QueryEscape(query))
	
	resp, err := s.client.R().
		SetContext(ctx).
		Get(searchURL)
	
	if err != nil {
		return nil, fmt.Errorf("reddit request failed: %w", err)
	}
	
	html := resp.String()
	results := make([]models.TavilyResult, 0, limit)
	
	// Parse posts
	postPattern := regexp.MustCompile(`<a class="search-title[^"]*" href="([^"]+)">([^<]+)</a>`)
	matches := postPattern.FindAllStringSubmatch(html, -1)
	
	for i := 0; i < len(matches) && i < limit; i++ {
		if len(matches[i]) < 3 {
			continue
		}
		
		postURL := matches[i][1]
		title := matches[i][2]
		
		// Get full post URL
		if strings.HasPrefix(postURL, "/r/") {
			postURL = "https://old.reddit.com" + postURL
		}
		
		results = append(results, models.TavilyResult{
			Title:   fmt.Sprintf("Reddit: %s", title),
			URL:     postURL,
			Content: title,
			Score:   0.8 - float64(i)*0.05,
		})
	}
	
	log.Printf("✅ Found %d Reddit results", len(results))
	return results, nil
}

// Habr scraping
func (s *SocialScraper) SearchHabr(ctx context.Context, query string, limit int) ([]models.TavilyResult, error) {
	log.Printf("🔍 Scraping Habr for: %s", query)
	
	searchURL := fmt.Sprintf("https://habr.com/ru/search/?q=%s&target_type=posts", 
		url.QueryEscape(query))
	
	resp, err := s.client.R().
		SetContext(ctx).
		Get(searchURL)
	
	if err != nil {
		return nil, fmt.Errorf("habr request failed: %w", err)
	}
	
	html := resp.String()
	results := make([]models.TavilyResult, 0, limit)
	
	// Parse articles
	titlePattern := regexp.MustCompile(`<a[^>]+class="tm-title__link"[^>]+href="([^"]+)"[^>]*><span>([^<]+)</span>`)
	snippetPattern := regexp.MustCompile(`<div class="article-formatted-body[^>]*>([^<]+)</div>`)
	
	titleMatches := titlePattern.FindAllStringSubmatch(html, -1)
	snippetMatches := snippetPattern.FindAllStringSubmatch(html, -1)
	
	for i := 0; i < len(titleMatches) && i < limit; i++ {
		if len(titleMatches[i]) < 3 {
			continue
		}
		
		articleURL := titleMatches[i][1]
		title := titleMatches[i][2]
		
		if !strings.HasPrefix(articleURL, "http") {
			articleURL = "https://habr.com" + articleURL
		}
		
		snippet := title
		if i < len(snippetMatches) && len(snippetMatches[i]) > 1 {
			snippet = snippetMatches[i][1]
			if len(snippet) > 200 {
				snippet = snippet[:200]
			}
		}
		
		results = append(results, models.TavilyResult{
			Title:   title,
			URL:     articleURL,
			Content: snippet,
			Score:   0.85 - float64(i)*0.05,
		})
	}
	
	log.Printf("✅ Found %d Habr results", len(results))
	return results, nil
}

// X/Twitter scraping (limited without API)
func (s *SocialScraper) SearchTwitter(ctx context.Context, query string, limit int) ([]models.TavilyResult, error) {
	log.Printf("🔍 Scraping Nitter (Twitter mirror) for: %s", query)
	
	// Use Nitter instance (Twitter frontend without JS)
	searchURL := fmt.Sprintf("https://nitter.net/search?q=%s", url.QueryEscape(query))
	
	resp, err := s.client.R().
		SetContext(ctx).
		Get(searchURL)
	
	if err != nil {
		return nil, fmt.Errorf("nitter request failed: %w", err)
	}
	
	html := resp.String()
	results := make([]models.TavilyResult, 0, limit)
	
	// Parse tweets
	tweetPattern := regexp.MustCompile(`<div class="tweet-content[^>]*>([^<]+)</div>`)
	matches := tweetPattern.FindAllStringSubmatch(html, -1)
	
	for i := 0; i < len(matches) && i < limit; i++ {
		if len(matches[i]) < 2 {
			continue
		}
		
		content := matches[i][1]
		if len(content) > 200 {
			content = content[:200]
		}
		
		results = append(results, models.TavilyResult{
			Title:   fmt.Sprintf("Twitter discussion: %s", truncate(content, 50)),
			URL:     searchURL,
			Content: content,
			Score:   0.7 - float64(i)*0.05,
		})
	}
	
	log.Printf("✅ Found %d Twitter results", len(results))
	return results, nil
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}