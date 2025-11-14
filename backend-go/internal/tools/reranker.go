package tools

import (
	"math"
	"strings"
	"unicode"

	"github.com/Vova4o/bootCamp2025CaseSber/backend/internal/models"
)

type BM25Reranker struct {
	k1 float64 // term frequency saturation parameter (обычно 1.2-2.0)
	b  float64 // length normalization parameter (обычно 0.75)
}

func NewBM25Reranker() *BM25Reranker {
	return &BM25Reranker{
		k1: 1.5,
		b:  0.75,
	}
}

// Rerank переранжирует результаты по релевантности к запросу
func (r *BM25Reranker) Rerank(
	query string,
	results []models.TavilyResult,
) []models.TavilyResult {
	if len(results) == 0 {
		return results
	}

	// Токенизация запроса
	queryTerms := r.tokenize(query)
	if len(queryTerms) == 0 {
		return results
	}

	// Подготовка документов
	docs := make([][]string, len(results))
	totalLen := 0
	for i, result := range results {
		text := result.Title + " " + result.Content
		docs[i] = r.tokenize(text)
		totalLen += len(docs[i])
	}

	avgDocLen := float64(totalLen) / float64(len(docs))

	// Вычисление IDF для каждого терма запроса
	idf := r.computeIDF(queryTerms, docs)

	// Вычисление BM25 score для каждого документа
	for i := range results {
		score := r.computeBM25(queryTerms, docs[i], avgDocLen, idf)
		results[i].Score = score
	}

	// Сортировка по score (descending)
	for i := 0; i < len(results)-1; i++ {
		for j := i + 1; j < len(results); j++ {
			if results[j].Score > results[i].Score {
				results[i], results[j] = results[j], results[i]
			}
		}
	}

	return results
}

// tokenize разбивает текст на токены
func (r *BM25Reranker) tokenize(text string) []string {
	text = strings.ToLower(text)
	
	// Разбиваем по пробелам и знакам препинания
	var tokens []string
	var current strings.Builder
	
	for _, r := range text {
		if unicode.IsLetter(r) || unicode.IsNumber(r) {
			current.WriteRune(r)
		} else if current.Len() > 0 {
			token := current.String()
			if len(token) > 2 { // Фильтруем короткие слова
				tokens = append(tokens, token)
			}
			current.Reset()
		}
	}
	
	if current.Len() > 0 {
		token := current.String()
		if len(token) > 2 {
			tokens = append(tokens, token)
		}
	}
	
	return tokens
}

// computeIDF вычисляет Inverse Document Frequency
func (r *BM25Reranker) computeIDF(
	queryTerms []string,
	docs [][]string,
) map[string]float64 {
	idf := make(map[string]float64)
	N := float64(len(docs))

	for _, term := range queryTerms {
		// Считаем в скольких документах встречается терм
		docCount := 0
		for _, doc := range docs {
			if r.contains(doc, term) {
				docCount++
			}
		}

		if docCount > 0 {
			// IDF = log((N - df + 0.5) / (df + 0.5) + 1)
			idf[term] = math.Log((N-float64(docCount)+0.5)/(float64(docCount)+0.5) + 1.0)
		}
	}

	return idf
}

// computeBM25 вычисляет BM25 score для документа
func (r *BM25Reranker) computeBM25(
	queryTerms []string,
	doc []string,
	avgDocLen float64,
	idf map[string]float64,
) float64 {
	score := 0.0
	docLen := float64(len(doc))

	// Подсчет частот термов в документе
	termFreq := make(map[string]int)
	for _, term := range doc {
		termFreq[term]++
	}

	// BM25 формула
	for _, term := range queryTerms {
		if idfScore, exists := idf[term]; exists {
			tf := float64(termFreq[term])
			
			// BM25 = IDF(term) * (tf * (k1 + 1)) / (tf + k1 * (1 - b + b * docLen / avgDocLen))
			numerator := tf * (r.k1 + 1)
			denominator := tf + r.k1*(1-r.b+r.b*docLen/avgDocLen)
			
			score += idfScore * (numerator / denominator)
		}
	}

	return score
}

// contains проверяет наличие терма в документе
func (r *BM25Reranker) contains(doc []string, term string) bool {
	for _, t := range doc {
		if t == term {
			return true
		}
	}
	return false
}