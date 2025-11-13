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

type FinanceScraper struct {
	client *resty.Client
}

func NewFinanceScraper() *FinanceScraper {
	client := resty.New()
	client.SetTimeout(15 * time.Second)
	return &FinanceScraper{client: client}
}

// Yahoo Finance scraping
func (s *FinanceScraper) SearchYahooFinance(ctx context.Context, query string, limit int) ([]models.TavilyResult, error) {
	log.Printf("🔍 Scraping Yahoo Finance for: %s", query)
	
	searchURL := fmt.Sprintf("https://finance.yahoo.com/search?q=%s", 
		url.QueryEscape(query))
	
	resp, err := s.client.R().
		SetContext(ctx).
		SetHeader("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7)").
		Get(searchURL)
	
	if err != nil {
		return nil, fmt.Errorf("yahoo finance request failed: %w", err)
	}
	
	html := resp.String()
	results := make([]models.TavilyResult, 0, limit)
	
	// Parse news articles
	newsPattern := regexp.MustCompile(`<a[^>]+data-test="quoteNews"[^>]+href="([^"]+)"[^>]*>([^<]+)</a>`)
	matches := newsPattern.FindAllStringSubmatch(html, -1)
	
	for i := 0; i < len(matches) && i < limit; i++ {
		if len(matches[i]) < 3 {
			continue
		}
		
		articleURL := matches[i][1]
		title := matches[i][2]
		
		if !strings.HasPrefix(articleURL, "http") {
			articleURL = "https://finance.yahoo.com" + articleURL
		}
		
		results = append(results, models.TavilyResult{
			Title:   fmt.Sprintf("[Yahoo Finance] %s", title),
			URL:     articleURL,
			Content: title,
			Score:   0.85 - float64(i)*0.05,
		})
	}
	
	log.Printf("✅ Found %d Yahoo Finance results", len(results))
	return results, nil
}

// Investing.com scraping
func (s *FinanceScraper) SearchInvestingCom(ctx context.Context, query string, limit int) ([]models.TavilyResult, error) {
	log.Printf("🔍 Scraping Investing.com for: %s", query)
	
	searchURL := fmt.Sprintf("https://www.investing.com/search/?q=%s", 
		url.QueryEscape(query))
	
	resp, err := s.client.R().
		SetContext(ctx).
		SetHeader("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64)").
		Get(searchURL)
	
	if err != nil {
		return nil, fmt.Errorf("investing.com request failed: %w", err)
	}
	
	html := resp.String()
	results := make([]models.TavilyResult, 0, limit)
	
	// Parse search results
	resultPattern := regexp.MustCompile(`<a[^>]+class="js-inner-all-results-quote-item"[^>]+href="([^"]+)"[^>]*>([^<]+)</a>`)
	matches := resultPattern.FindAllStringSubmatch(html, -1)
	
	for i := 0; i < len(matches) && i < limit; i++ {
		if len(matches[i]) < 3 {
			continue
		}
		
		articleURL := matches[i][1]
		title := matches[i][2]
		
		if !strings.HasPrefix(articleURL, "http") {
			articleURL = "https://www.investing.com" + articleURL
		}
		
		results = append(results, models.TavilyResult{
			Title:   title,
			URL:     articleURL,
			Content: title,
			Score:   0.8 - float64(i)*0.05,
		})
	}
	
	log.Printf("✅ Found %d Investing.com results", len(results))
	return results, nil
}

// MarketWatch scraping
func (s *FinanceScraper) SearchMarketWatch(ctx context.Context, query string, limit int) ([]models.TavilyResult, error) {
	log.Printf("🔍 Scraping MarketWatch for: %s", query)
	
	searchURL := fmt.Sprintf("https://www.marketwatch.com/search?q=%s", 
		url.QueryEscape(query))
	
	resp, err := s.client.R().
		SetContext(ctx).
		Get(searchURL)
	
	if err != nil {
		return nil, fmt.Errorf("marketwatch request failed: %w", err)
	}
	
	html := resp.String()
	results := make([]models.TavilyResult, 0, limit)
	
	// Parse articles
	articlePattern := regexp.MustCompile(`<h3[^>]*><a[^>]+href="([^"]+)"[^>]*>([^<]+)</a>`)
	matches := articlePattern.FindAllStringSubmatch(html, -1)
	
	for i := 0; i < len(matches) && i < limit; i++ {
		if len(matches[i]) < 3 {
			continue
		}
		
		articleURL := matches[i][1]
		title := matches[i][2]
		
		results = append(results, models.TavilyResult{
			Title:   fmt.Sprintf("[MarketWatch] %s", title),
			URL:     articleURL,
			Content: title,
			Score:   0.85 - float64(i)*0.05,
		})
	}
	
	log.Printf("✅ Found %d MarketWatch results", len(results))
	return results, nil
}