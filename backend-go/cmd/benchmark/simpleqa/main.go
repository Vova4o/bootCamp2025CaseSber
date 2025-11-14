package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

// ============================================================================
// ДОБАВЛЕНО: Типы для API запросов/ответов
// ============================================================================

type SearchRequest struct {
	Query string `json:"query"`
	Mode  string `json:"mode"`
}

type SearchResponse struct {
	Answer  string   `json:"answer"`
	Sources []Source `json:"sources"`
}

type Source struct {
	Title   string  `json:"title"`
	URL     string  `json:"url"`
	Snippet string  `json:"snippet"`
	Score   float64 `json:"score,omitempty"`
}

// ============================================================================
// Dataset и результаты
// ============================================================================

type SimpleQAQuestion struct {
	Question       string   `json:"question"`
	Answer         string   `json:"answer"`
	Category       string   `json:"category"`
	AcceptableVars []string `json:"acceptable_variations"` // ДОБАВЛЕНО: варианты ответа
}

type BenchmarkResult struct {
	Question       string        `json:"question"`
	ExpectedAnswer string        `json:"expected_answer"`
	ActualAnswer   string        `json:"actual_answer"`
	Category       string        `json:"category"`
	ProcessingTime time.Duration `json:"processing_time"`
	Correct        bool          `json:"correct"`        // ИЗМЕНЕНО: вместо Success
	HasSources     bool          `json:"has_sources"`    // ДОБАВЛЕНО
	SourceCount    int           `json:"source_count"`   // ДОБАВЛЕНО
	Mode           string        `json:"mode"`
}

// ДОБАВЛЕНО: Статистика по категориям
type CategoryStats struct {
	Total    int
	Correct  int
	Accuracy float64
}

type Stats struct {
	TotalQuestions int
	CorrectCount   int              // ИЗМЕНЕНО: вместо SuccessCount
	FailCount      int
	Accuracy       float64          // ДОБАВЛЕНО: вместо SuccessRate
	AvgTime        float64
	TotalTime      time.Duration
	ByCategory     map[string]CategoryStats // ДОБАВЛЕНО
}

func main() {
	mode := flag.String("mode", "simple", "Mode to test: simple or pro")
	dataFile := flag.String("data", "simpleqa_dataset.json", "Path to SimpleQA dataset JSON file")
	limit := flag.Int("limit", 10, "Number of questions to test (0 = all)")
	output := flag.String("output", "benchmark_results.json", "Output file for results")
	apiURL := flag.String("api", "http://localhost:8000", "Backend API URL")
	flag.Parse()

	log.Printf("🧪 SimpleQA Benchmark - Mode: %s, API: %s", *mode, *apiURL)

	questions, err := loadDataset(*dataFile)
	if err != nil {
		log.Fatalf("Failed to load dataset: %v", err)
	}

	log.Printf("Loaded %d questions from dataset", len(questions))

	if *limit > 0 && *limit < len(questions) {
		questions = questions[:*limit]
	}

	results := make([]BenchmarkResult, 0, len(questions))
	startTime := time.Now()

	for i, q := range questions {
		log.Printf("\n[%d/%d] ❓ %s", i+1, len(questions), q.Question)
		log.Printf("  📌 Expected: %s", q.Answer)

		result := runQuestion(*apiURL, q, *mode)
		results = append(results, result)

		status := "✅"
		if !result.Correct {
			status = "❌"
		}

		actualAnswer := result.ActualAnswer
		if len(actualAnswer) > 150 {
			actualAnswer = actualAnswer[:150] + "..."
		}
		log.Printf("  💬 Got: %s", actualAnswer)
		log.Printf("  %s %s (%.2fs, %d sources)", 
			status, 
			map[bool]string{true: "CORRECT", false: "INCORRECT"}[result.Correct],
			result.ProcessingTime.Seconds(),
			result.SourceCount)
	}

	totalTime := time.Since(startTime)
	stats := calculateStats(results, totalTime)
	printSummary(stats, *mode)

	if err := saveResults(results, *output); err != nil {
		log.Printf("Warning: Failed to save results: %v", err)
	} else {
		log.Printf("Results saved to %s", *output)
	}
}

func loadDataset(filename string) ([]SimpleQAQuestion, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		if os.IsNotExist(err) {
			log.Printf("Dataset not found, creating sample dataset...")
			return createSampleDataset(), nil
		}
		return nil, err
	}

	var questions []SimpleQAQuestion
	if err := json.Unmarshal(data, &questions); err != nil {
		return nil, err
	}
	return questions, nil
}

// УЛУЧШЕНО: Добавлены acceptable_variations для каждого вопроса
func createSampleDataset() []SimpleQAQuestion {
	return []SimpleQAQuestion{
		{
			Question: "What is the capital of France?",
			Answer:   "Paris",
			Category: "geography",
			AcceptableVars: []string{"paris", "Paris"},
		},
		{
			Question: "Who wrote Romeo and Juliet?",
			Answer:   "William Shakespeare",
			Category: "literature",
			AcceptableVars: []string{"shakespeare", "William Shakespeare", "Shakespeare"},
		},
		{
			Question: "What is the largest planet in our solar system?",
			Answer:   "Jupiter",
			Category: "astronomy",
			AcceptableVars: []string{"jupiter", "Jupiter"},
		},
		{
			Question: "In what year did World War II end?",
			Answer:   "1945",
			Category: "history",
			AcceptableVars: []string{"1945"},
		},
		{
			Question: "What is the speed of light?",
			Answer:   "299,792,458 m/s",
			Category: "physics",
			AcceptableVars: []string{"299792458", "300000000", "3*10^8", "299,792,458", "approximately 300,000"},
		},
		{
			Question: "Who painted the Mona Lisa?",
			Answer:   "Leonardo da Vinci",
			Category: "art",
			AcceptableVars: []string{"da vinci", "leonardo", "Leonardo da Vinci"},
		},
		{
			Question: "What is the chemical symbol for gold?",
			Answer:   "Au",
			Category: "chemistry",
			AcceptableVars: []string{"au", "Au", "AU"},
		},
		{
			Question: "How many continents are there?",
			Answer:   "7",
			Category: "geography",
			AcceptableVars: []string{"7", "seven", "Seven"},
		},
		{
			Question: "What is the largest ocean?",
			Answer:   "Pacific Ocean",
			Category: "geography",
			AcceptableVars: []string{"pacific", "Pacific", "Pacific Ocean"},
		},
		{
			Question: "Who invented the telephone?",
			Answer:   "Alexander Graham Bell",
			Category: "history",
			AcceptableVars: []string{"bell", "graham bell", "Alexander Graham Bell", "Alexander Bell"},
		},
	}
}

func runQuestion(apiURL string, q SimpleQAQuestion, mode string) BenchmarkResult {
	start := time.Now()

	reqBody := SearchRequest{
		Query: q.Question,
		Mode:  mode,
	}
	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return BenchmarkResult{
			Question:       q.Question,
			ExpectedAnswer: q.Answer,
			ActualAnswer:   fmt.Sprintf("ERROR marshaling request: %v", err),
			Category:       q.Category,
			ProcessingTime: time.Since(start),
			Correct:        false,
			Mode:           mode,
		}
	}

	resp, err := http.Post(apiURL+"/api/search", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return BenchmarkResult{
			Question:       q.Question,
			ExpectedAnswer: q.Answer,
			ActualAnswer:   fmt.Sprintf("ERROR calling API: %v", err),
			Category:       q.Category,
			ProcessingTime: time.Since(start),
			Correct:        false,
			Mode:           mode,
		}
	}
	defer resp.Body.Close()

	processingTime := time.Since(start)

	if resp.StatusCode != http.StatusOK {
		return BenchmarkResult{
			Question:       q.Question,
			ExpectedAnswer: q.Answer,
			ActualAnswer:   fmt.Sprintf("ERROR: HTTP %d", resp.StatusCode),
			Category:       q.Category,
			ProcessingTime: processingTime,
			Correct:        false,
			Mode:           mode,
		}
	}

	var result SearchResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return BenchmarkResult{
			Question:       q.Question,
			ExpectedAnswer: q.Answer,
			ActualAnswer:   fmt.Sprintf("ERROR parsing response: %v", err),
			Category:       q.Category,
			ProcessingTime: processingTime,
			Correct:        false,
			Mode:           mode,
		}
	}

	// ДОБАВЛЕНО: Проверка корректности ответа
	correct := checkAnswer(result.Answer, q.Answer, q.AcceptableVars)

	return BenchmarkResult{
		Question:       q.Question,
		ExpectedAnswer: q.Answer,
		ActualAnswer:   result.Answer,
		Category:       q.Category,
		ProcessingTime: processingTime,
		Correct:        correct,              // ИЗМЕНЕНО
		HasSources:     len(result.Sources) > 0,
		SourceCount:    len(result.Sources),
		Mode:           mode,
	}
}

// ДОБАВЛЕНО: Функция проверки правильности ответа
func checkAnswer(actual, expected string, acceptable []string) bool {
	if actual == "" {
		return false
	}

	actualLower := strings.ToLower(strings.TrimSpace(actual))
	
	// Проверяем точное совпадение с ожидаемым ответом
	if strings.Contains(actualLower, strings.ToLower(expected)) {
		return true
	}

	// Проверяем варианты
	for _, variant := range acceptable {
		if strings.Contains(actualLower, strings.ToLower(variant)) {
			return true
		}
	}

	return false
}

// УЛУЧШЕНО: Добавлена статистика по категориям
func calculateStats(results []BenchmarkResult, totalTime time.Duration) Stats {
	stats := Stats{
		TotalQuestions: len(results),
		TotalTime:      totalTime,
		ByCategory:     make(map[string]CategoryStats),
	}

	var totalProcessingTime time.Duration
	for _, r := range results {
		if r.Correct {
			stats.CorrectCount++
		} else {
			stats.FailCount++
		}
		totalProcessingTime += r.ProcessingTime

		// Статистика по категориям
		catStats := stats.ByCategory[r.Category]
		catStats.Total++
		if r.Correct {
			catStats.Correct++
		}
		stats.ByCategory[r.Category] = catStats
	}

	if stats.TotalQuestions > 0 {
		stats.Accuracy = float64(stats.CorrectCount) / float64(stats.TotalQuestions) * 100
		stats.AvgTime = totalProcessingTime.Seconds() / float64(stats.TotalQuestions)
	}

	// Вычисляем accuracy по категориям
	for cat, catStats := range stats.ByCategory {
		catStats.Accuracy = float64(catStats.Correct) / float64(catStats.Total) * 100
		stats.ByCategory[cat] = catStats
	}

	return stats
}

// УЛУЧШЕНО: Более красивый вывод с категориями
func printSummary(stats Stats, mode string) {
	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Printf("      SimpleQA BENCHMARK RESULTS (%s Mode)\n", strings.ToUpper(mode))
	fmt.Println(strings.Repeat("=", 60))

	fmt.Printf("\n📊 Overall Statistics:\n")
	fmt.Printf("  Total Questions: %d\n", stats.TotalQuestions)
	fmt.Printf("  ✅ Correct: %d\n", stats.CorrectCount)
	fmt.Printf("  ❌ Incorrect: %d\n", stats.FailCount)
	fmt.Printf("  🎯 Accuracy: %.2f%%\n", stats.Accuracy)

	fmt.Printf("\n📚 By Category:\n")
	for cat, catStats := range stats.ByCategory {
		icon := "✅"
		if catStats.Accuracy < 50 {
			icon = "❌"
		} else if catStats.Accuracy < 80 {
			icon = "⚠️"
		}
		fmt.Printf("  %s %s: %.1f%% (%d/%d)\n",
			icon, cat, catStats.Accuracy, catStats.Correct, catStats.Total)
	}

	fmt.Printf("\n⏱️  Performance:\n")
	fmt.Printf("  Average Time: %.2fs per question\n", stats.AvgTime)
	fmt.Printf("  Total Time: %.2fs\n", stats.TotalTime.Seconds())

	fmt.Println("\n" + strings.Repeat("=", 60))
}

func saveResults(results []BenchmarkResult, filename string) error {
	data, err := json.MarshalIndent(results, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filename, data, 0o644)
}