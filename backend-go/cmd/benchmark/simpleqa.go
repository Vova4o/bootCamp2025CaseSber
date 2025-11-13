package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/Vova4o/bootCamp2025CaseSber/backend/internal/agents"
	"github.com/Vova4o/bootCamp2025CaseSber/backend/internal/config"
)

type SimpleQAQuestion struct {
	Question string `json:"question"`
	Answer   string `json:"answer"`
	Category string `json:"category"`
}

type BenchmarkResult struct {
	Question       string        `json:"question"`
	ExpectedAnswer string        `json:"expected_answer"`
	ActualAnswer   string        `json:"actual_answer"`
	Category       string        `json:"category"`
	ProcessingTime time.Duration `json:"processing_time"`
	Success        bool          `json:"success"`
	Mode           string        `json:"mode"`
}

type Stats struct {
	TotalQuestions int
	SuccessCount   int
	FailCount      int
	SuccessRate    float64
	AvgTime        float64
	TotalTime      time.Duration
}

func main() {
	mode := flag.String("mode", "simple", "Mode to test: simple, pro, or auto")
	dataFile := flag.String("data", "simpleqa_dataset.json", "Path to SimpleQA dataset JSON file")
	limit := flag.Int("limit", 10, "Number of questions to test (0 = all)")
	output := flag.String("output", "benchmark_results.json", "Output file for results")
	flag.Parse()

	cfg := config.LoadConfig()
	router := agents.NewRouterAgent(cfg)

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
		log.Printf("[%d/%d] Testing: %s", i+1, len(questions), q.Question)
		result := runQuestion(router, q, *mode)
		results = append(results, result)

		status := "✅"
		if !result.Success {
			status = "❌"
		}
		log.Printf("  %s (%.2fs)", status, result.ProcessingTime.Seconds())
	}

	totalTime := time.Since(startTime)
	stats := calculateStats(results, totalTime)
	printSummary(stats)

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

func createSampleDataset() []SimpleQAQuestion {
	return []SimpleQAQuestion{
		{Question: "What is the capital of France?", Answer: "Paris", Category: "geography"},
		{Question: "Who wrote Romeo and Juliet?", Answer: "William Shakespeare", Category: "literature"},
		{Question: "What is the largest planet in our solar system?", Answer: "Jupiter", Category: "science"},
		{Question: "In what year did World War II end?", Answer: "1945", Category: "history"},
		{Question: "What is the speed of light?", Answer: "299,792,458 meters per second", Category: "science"},
		{Question: "Who painted the Mona Lisa?", Answer: "Leonardo da Vinci", Category: "art"},
		{Question: "What is the chemical symbol for gold?", Answer: "Au", Category: "science"},
		{Question: "How many continents are there?", Answer: "7", Category: "geography"},
		{Question: "What is the largest ocean?", Answer: "Pacific Ocean", Category: "geography"},
		{Question: "Who invented the telephone?", Answer: "Alexander Graham Bell", Category: "history"},
	}
}

func runQuestion(router *agents.RouterAgent, q SimpleQAQuestion, mode string) BenchmarkResult {
	ctx := context.Background()
	start := time.Now()

	result, err := router.ProcessQuery(ctx, q.Question, mode)
	processingTime := time.Since(start)

	if err != nil {
		return BenchmarkResult{
			Question:       q.Question,
			ExpectedAnswer: q.Answer,
			ActualAnswer:   fmt.Sprintf("ERROR: %v", err),
			Category:       q.Category,
			ProcessingTime: processingTime,
			Success:        false,
			Mode:           mode,
		}
	}

	success := result.Answer != "" && len(result.Sources) > 0

	return BenchmarkResult{
		Question:       q.Question,
		ExpectedAnswer: q.Answer,
		ActualAnswer:   result.Answer,
		Category:       q.Category,
		ProcessingTime: processingTime,
		Success:        success,
		Mode:           mode,
	}
}

func calculateStats(results []BenchmarkResult, totalTime time.Duration) Stats {
	stats := Stats{
		TotalQuestions: len(results),
		TotalTime:      totalTime,
	}

	var totalProcessingTime time.Duration
	for _, r := range results {
		if r.Success {
			stats.SuccessCount++
		} else {
			stats.FailCount++
		}
		totalProcessingTime += r.ProcessingTime
	}

	if stats.TotalQuestions > 0 {
		stats.SuccessRate = float64(stats.SuccessCount) / float64(stats.TotalQuestions) * 100
		stats.AvgTime = totalProcessingTime.Seconds() / float64(stats.TotalQuestions)
	}

	return stats
}

func printSummary(stats Stats) {
	fmt.Println("\n========== BENCHMARK RESULTS ==========")
	fmt.Printf("\n📊 Statistics:\n")
	fmt.Printf("  Total Questions: %d\n", stats.TotalQuestions)
	fmt.Printf("  ✅ Success: %d\n", stats.SuccessCount)
	fmt.Printf("  ❌ Failed: %d\n", stats.FailCount)
	fmt.Printf("  Success Rate: %.2f%%\n", stats.SuccessRate)
	fmt.Printf("\n⏱️  Performance:\n")
	fmt.Printf("  Average Time: %.2fs per question\n", stats.AvgTime)
	fmt.Printf("  Total Time: %.2fs\n", stats.TotalTime.Seconds())
	fmt.Println("\n=======================================")
}

func saveResults(results []BenchmarkResult, filename string) error {
	data, err := json.MarshalIndent(results, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filename, data, 0o644)
}
