package tools

import (
	"net/url"
	"strings"
	"time"

	"github.com/Vova4o/bootCamp2025CaseSber/backend/internal/models"
)

type CredibilityScorer struct{}

func NewCredibilityScorer() *CredibilityScorer {
	return &CredibilityScorer{}
}

// ScoreSource оценивает достоверность источника (0.0 - 1.0)
func (c *CredibilityScorer) ScoreSource(source models.TavilyResult) float64 {
	score := 0.5 // базовый score

	// 1. Domain authority (30% веса)
	domainScore := c.scoreDomain(source.URL)
	score += domainScore * 0.3

	// 2. Content quality (25% веса)
	contentScore := c.scoreContent(source.Content, source.Title)
	score += contentScore * 0.25

	// 3. Relevance score from search (25% веса)
	score += source.Score * 0.25

	// 4. URL quality (10% веса)
	urlScore := c.scoreURL(source.URL)
	score += urlScore * 0.1

	// 5. Freshness (10% веса)
	freshnessScore := c.scoreFreshness(source.URL)
	score += freshnessScore * 0.1

	// Нормализация в диапазон 0-1
	if score > 1.0 {
		score = 1.0
	}
	if score < 0.0 {
		score = 0.0
	}

	return score
}

// scoreDomain оценивает надежность домена
func (c *CredibilityScorer) scoreDomain(urlStr string) float64 {
	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		return 0.3
	}

	domain := strings.ToLower(parsedURL.Hostname())

	// Высоконадежные домены (0.9-1.0)
	highTrustDomains := []string{
		"wikipedia.org", "wikimedia.org",
		".gov", ".edu",
		"nature.com", "science.org", "sciencedirect.com",
		"nih.gov", "cdc.gov",
		"bbc.com", "reuters.com", "apnews.com",
		"arxiv.org", "scholar.google.com",
		"nist.gov", "ieee.org", "acm.org",
	}

	for _, trusted := range highTrustDomains {
		if strings.Contains(domain, trusted) {
			return 1.0
		}
	}

	// Среднена дежные домены (0.7-0.8)
	mediumTrustDomains := []string{
		".org", "github.com", "stackoverflow.com",
		"medium.com", "habr.com", "vc.ru",
		"forbes.com", "techcrunch.com", "theverge.com",
		"nytimes.com", "theguardian.com", "washingtonpost.com",
	}

	for _, medium := range mediumTrustDomains {
		if strings.Contains(domain, medium) {
			return 0.75
		}
	}

	// Блоги и личные сайты (0.4-0.6)
	if strings.Contains(domain, "blog") ||
		strings.Contains(domain, "wordpress") ||
		strings.Contains(domain, "blogspot") {
		return 0.5
	}

	// Социальные сети (0.3-0.5)
	socialDomains := []string{
		"facebook.com", "twitter.com", "x.com",
		"reddit.com", "quora.com",
		"vk.com", "ok.ru",
	}

	for _, social := range socialDomains {
		if strings.Contains(domain, social) {
			return 0.4
		}
	}

	// Неизвестные домены
	return 0.5
}

// scoreContent оценивает качество контента
func (c *CredibilityScorer) scoreContent(content, title string) float64 {
	score := 0.5

	// Длина контента
	contentLen := len(content)
	if contentLen > 500 {
		score += 0.2
	} else if contentLen > 200 {
		score += 0.1
	}

	// Наличие структурированной информации
	structureKeywords := []string{
		"источник", "исследование", "данные", "статистика",
		"study", "research", "data", "source", "published",
		"согласно", "по данным", "according to",
	}

	for _, keyword := range structureKeywords {
		if strings.Contains(strings.ToLower(content), keyword) {
			score += 0.05
			break
		}
	}

	// Наличие дат
	datePatterns := []string{
		"202", "201", // годы
		"января", "февраля", "марта", "april", "may", "june",
	}

	for _, pattern := range datePatterns {
		if strings.Contains(strings.ToLower(content), pattern) {
			score += 0.05
			break
		}
	}

	// Избегаем кликбейта
	clickbaitWords := []string{
		"невероятно", "шокирующ", "сенсаци", "тайн",
		"shocking", "incredible", "secret", "mystery",
		"🔥", "😱", "!!!",
	}

	for _, clickbait := range clickbaitWords {
		if strings.Contains(strings.ToLower(title), clickbait) {
			score -= 0.1
			break
		}
	}

	if score > 1.0 {
		score = 1.0
	}
	if score < 0.0 {
		score = 0.0
	}

	return score
}

// scoreURL оценивает качество URL
func (c *CredibilityScorer) scoreURL(urlStr string) float64 {
	score := 0.5

	// HTTPS
	if strings.HasPrefix(urlStr, "https://") {
		score += 0.2
	}

	// Длина URL (короткие URL лучше)
	if len(urlStr) < 100 {
		score += 0.2
	} else if len(urlStr) > 200 {
		score -= 0.1
	}

	// Подозрительные паттерны
	suspiciousPatterns := []string{
		"bit.ly", "tinyurl", "goo.gl", // сокращенные URL
		"?ref=", "?utm_", // tracking параметры (много)
		"ad", "promo", // рекламные страницы
	}

	for _, pattern := range suspiciousPatterns {
		if strings.Contains(strings.ToLower(urlStr), pattern) {
			score -= 0.1
		}
	}

	if score > 1.0 {
		score = 1.0
	}
	if score < 0.0 {
		score = 0.0
	}

	return score
}

// scoreFreshness оценивает свежесть контента (если можно определить)
func (c *CredibilityScorer) scoreFreshness(urlStr string) float64 {
	// Простая эвристика - проверяем наличие года в URL
	currentYear := time.Now().Year()
	
	for year := currentYear; year >= currentYear-5; year-- {
		if strings.Contains(urlStr, string(rune(year))) {
			yearsOld := currentYear - year
			// Свежие источники (0-2 года) = 1.0
			// Старые (3-5 лет) = 0.5-0.8
			if yearsOld <= 2 {
				return 1.0
			}
			return 1.0 - float64(yearsOld)*0.1
		}
	}

	return 0.5 // Не удалось определить
}

// RankSources сортирует источники по credibility
func (c *CredibilityScorer) RankSources(sources []models.TavilyResult) []models.TavilyResult {
	for i := range sources {
		sources[i].Credibility = c.ScoreSource(sources[i])
	}

	// Сортировка по credibility (descending)
	for i := 0; i < len(sources)-1; i++ {
		for j := i + 1; j < len(sources); j++ {
			if sources[j].Credibility > sources[i].Credibility {
				sources[i], sources[j] = sources[j], sources[i]
			}
		}
	}

	return sources
}