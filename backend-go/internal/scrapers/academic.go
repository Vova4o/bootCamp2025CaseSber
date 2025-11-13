package scrapers

import (
	"context"
	"fmt"
	"log"
	"net/url"
	"strings"
	"time"

	"github.com/Vova4o/bootCamp2025CaseSber/backend/internal/models"
	"github.com/go-resty/resty/v2"
)

type AcademicScraper struct {
	client *resty.Client
}

func NewAcademicScraper() *AcademicScraper {
	client := resty.New()
	client.SetTimeout(15 * time.Second)
	return &AcademicScraper{client: client}
}

// arXiv API response
type ArxivResponse struct {
	Feed struct {
		Entry []struct {
			ID      string `json:"id"`
			Title   string `json:"title"`
			Summary string `json:"summary"`
			Link    []struct {
				Href string `json:"href"`
			} `json:"link"`
		} `json:"entry"`
	} `json:"feed"`
}

// Search arXiv
func (s *AcademicScraper) SearchArxiv(ctx context.Context, query string, limit int) ([]models.TavilyResult, error) {
	log.Printf("🔍 Searching arXiv for: %s", query)

	searchURL := fmt.Sprintf(
		"http://export.arxiv.org/api/query?search_query=all:%s&start=0&max_results=%d&sortBy=relevance&sortOrder=descending",
		url.QueryEscape(query), limit)

	resp, err := s.client.R().
		SetContext(ctx).
		SetHeader("Accept", "application/json").
		Get(searchURL)
	if err != nil {
		return nil, fmt.Errorf("arxiv request failed: %w", err)
	}

	// Parse XML response (arXiv returns Atom XML)
	xml := resp.String()
	results := make([]models.TavilyResult, 0, limit)

	// Simple regex parsing for entries
	entries := extractXMLEntries(xml)

	for i, entry := range entries {
		if i >= limit {
			break
		}

		results = append(results, models.TavilyResult{
			Title:   fmt.Sprintf("[arXiv] %s", entry.Title),
			URL:     entry.URL,
			Content: entry.Summary,
			Score:   0.95 - float64(i)*0.03,
		})
	}

	log.Printf("✅ Found %d arXiv papers", len(results))
	return results, nil
}

type XMLEntry struct {
	Title   string
	URL     string
	Summary string
}

func extractXMLEntries(xml string) []XMLEntry {
	// Simple extraction (in production use proper XML parser)
	entries := make([]XMLEntry, 0)

	// Split by <entry> tags
	parts := splitByTag(xml, "entry")

	for _, part := range parts {
		entry := XMLEntry{
			Title:   extractBetween(part, "<title>", "</title>"),
			URL:     extractBetween(part, `<id>`, `</id>`),
			Summary: extractBetween(part, "<summary>", "</summary>"),
		}

		if entry.Title != "" && entry.URL != "" {
			// Clean up
			entry.Title = cleanXMLText(entry.Title)
			entry.Summary = cleanXMLText(entry.Summary)
			if len(entry.Summary) > 300 {
				entry.Summary = entry.Summary[:300] + "..."
			}
			entries = append(entries, entry)
		}
	}

	return entries
}

// Google Scholar scraping (limited)
func (s *AcademicScraper) SearchGoogleScholar(ctx context.Context, query string, limit int) ([]models.TavilyResult, error) {
	log.Printf("🔍 Scraping Google Scholar for: %s", query)

	searchURL := fmt.Sprintf("https://scholar.google.com/scholar?q=%s&hl=en",
		url.QueryEscape(query))

	resp, err := s.client.R().
		SetContext(ctx).
		SetHeader("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36").
		Get(searchURL)
	if err != nil {
		return nil, fmt.Errorf("scholar request failed: %w", err)
	}

	html := resp.String()
	results := make([]models.TavilyResult, 0, limit)

	// Parse results (Google Scholar has specific structure)
	papers := parseScholarResults(html, limit)

	for i, paper := range papers {
		results = append(results, models.TavilyResult{
			Title:   fmt.Sprintf("[Scholar] %s", paper.Title),
			URL:     paper.URL,
			Content: paper.Snippet,
			Score:   0.9 - float64(i)*0.04,
		})
	}

	log.Printf("✅ Found %d Scholar papers", len(results))
	return results, nil
}

type ScholarPaper struct {
	Title   string
	URL     string
	Snippet string
}

func parseScholarResults(html string, limit int) []ScholarPaper {
	papers := make([]ScholarPaper, 0, limit)

	// Split by result divs
	parts := splitByTag(html, `<div class="gs_ri">`)

	for i := 1; i < len(parts) && len(papers) < limit; i++ {
		part := parts[i]

		title := extractBetween(part, `<h3`, `</h3>`)
		title = extractBetween(title, `>`, `<`)

		url := extractBetween(part, `href="`, `"`)

		snippet := extractBetween(part, `<div class="gs_rs">`, `</div>`)
		snippet = cleanXMLText(snippet)

		if title != "" {
			papers = append(papers, ScholarPaper{
				Title:   title,
				URL:     url,
				Snippet: snippet,
			})
		}
	}

	return papers
}

func splitByTag(text, tag string) []string {
	// Simple split by tag
	return []string{text} // Simplified
}

func extractBetween(text, start, end string) string {
	startIdx := 0
	if start != "" {
		idx := indexOf(text, start)
		if idx == -1 {
			return ""
		}
		startIdx = idx + len(start)
	}

	endIdx := len(text)
	if end != "" {
		idx := indexOf(text[startIdx:], end)
		if idx == -1 {
			return ""
		}
		endIdx = startIdx + idx
	}

	return text[startIdx:endIdx]
}

func indexOf(text, substr string) int {
	for i := 0; i <= len(text)-len(substr); i++ {
		if text[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}

func cleanXMLText(text string) string {
	text = stripHTMLTags(text)
	text = strings.TrimSpace(text)
	return text
}

func stripHTMLTags(text string) string {
	// Remove HTML tags
	result := ""
	inTag := false
	for _, char := range text {
		if char == '<' {
			inTag = true
		} else if char == '>' {
			inTag = false
		} else if !inTag {
			result += string(char)
		}
	}
	return result
}
