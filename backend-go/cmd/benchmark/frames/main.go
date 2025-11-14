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

type FRAMESQuestion struct {
	Question        string   `json:"question"`
	Answer          string   `json:"answer"`
	Category        string   `json:"category"`
	HopCount        int      `json:"hop_count"`
	RequiredSources int      `json:"required_sources"`
	Keywords        []string `json:"keywords"`
}

type FRAMESResult struct {
	Question        string        `json:"question"`
	ExpectedAnswer  string        `json:"expected_answer"`
	ActualAnswer    string        `json:"actual_answer"`
	Category        string        `json:"category"`
	ProcessingTime  time.Duration `json:"processing_time"`
	FactualityScore float64       `json:"factuality_score"`
	ReasoningDepth  float64       `json:"reasoning_depth"`
	SourceDiversity float64       `json:"source_diversity"`
	SourceCount     int           `json:"source_count"`
	HopCount        int           `json:"hop_count"`
	Success         bool          `json:"success"`
	Mode            string        `json:"mode"`
}

type FRAMESStats struct {
	TotalQuestions    int
	SuccessCount      int
	FailCount         int
	SuccessRate       float64
	AvgFactuality     float64
	AvgReasoningDepth float64
	AvgSourceDiv      float64
	AvgTime           float64
	TotalTime         time.Duration
}

func main() {
	mode := flag.String("mode", "pro", "Mode to test: pro")
	dataFile := flag.String("data", "frames_dataset.json", "Path to FRAMES dataset JSON file")
	limit := flag.Int("limit", 10, "Number of questions to test (0 = all)")
	output := flag.String("output", "frames_results.json", "Output file for results")
	apiURL := flag.String("api", "http://localhost:8000", "Backend API URL")
	flag.Parse()

	log.Printf("🧪 FRAMES Benchmark - Using API: %s", *apiURL)

	questions, err := loadFRAMESDataset(*dataFile)
	if err != nil {
		log.Fatalf("Failed to load dataset: %v", err)
	}

	log.Printf("Loaded %d multi-hop questions from dataset", len(questions))

	if *limit > 0 && *limit < len(questions) {
		questions = questions[:*limit]
	}

	results := make([]FRAMESResult, 0, len(questions))
	startTime := time.Now()

	for i, q := range questions {
		log.Printf("\n[%d/%d] 🔬 Multi-hop Question (%d hops): %s",
			i+1, len(questions), q.HopCount, q.Question)
		log.Printf("  📌 Expected: %s", q.Answer)
		log.Printf("  🔑 Keywords: %v", q.Keywords)

		result := runFRAMESQuestion(*apiURL, q, *mode)
		results = append(results, result)

		status := "✅"
		if !result.Success {
			status = "❌"
		}

		log.Printf("  💬 Got: %s", truncate(result.ActualAnswer, 150))
		log.Printf("  %s Scores: Factuality=%.2f, Depth=%.2f, Diversity=%.2f (%.2fs)",
			status, result.FactualityScore, result.ReasoningDepth,
			result.SourceDiversity, result.ProcessingTime.Seconds())
	}

	totalTime := time.Since(startTime)
	stats := calculateFRAMESStats(results, totalTime)
	printFRAMESSummary(stats)

	if err := saveFRAMESResults(results, *output); err != nil {
		log.Printf("Warning: Failed to save results: %v", err)
	} else {
		log.Printf("Results saved to %s", *output)
	}
}

func loadFRAMESDataset(filename string) ([]FRAMESQuestion, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		if os.IsNotExist(err) {
			log.Printf("Dataset not found, creating sample FRAMES dataset...")
			return createSampleFRAMESDataset(), nil
		}
		return nil, err
	}

	var questions []FRAMESQuestion
	if err := json.Unmarshal(data, &questions); err != nil {
		return nil, err
	}
	return questions, nil
}

func createSampleFRAMESDataset() []FRAMESQuestion {
	return []FRAMESQuestion{
		{
			Question:        "Compare the economic policies of the US and EU in response to the 2008 financial crisis",
			Answer:          "US focused on quantitative easing and bank bailouts, EU emphasized austerity measures",
			Category:        "economics",
			HopCount:        3,
			RequiredSources: 5,
			Keywords:        []string{"2008", "financial crisis", "US", "EU", "policy"},
		},
		{
			Question:        "How does climate change affect agricultural productivity in developing countries?",
			Answer:          "Increased droughts, floods, and temperature changes reduce crop yields",
			Category:        "climate",
			HopCount:        2,
			RequiredSources: 4,
			Keywords:        []string{"climate change", "agriculture", "developing countries"},
		},
		{
			Question:        "Explain the relationship between social media usage and mental health in teenagers",
			Answer:          "Studies show correlation with anxiety, depression, but causation is debated",
			Category:        "health",
			HopCount:        2,
			RequiredSources: 4,
			Keywords:        []string{"social media", "mental health", "teenagers"},
		},
		{
			Question:        "What are the advantages and disadvantages of nuclear energy compared to renewable sources?",
			Answer:          "Nuclear: reliable, low emissions but waste issues. Renewables: clean but intermittent",
			Category:        "energy",
			HopCount:        3,
			RequiredSources: 5,
			Keywords:        []string{"nuclear energy", "renewable", "advantages", "disadvantages"},
		},
		{
			Question:        "How did the invention of the printing press influence the Protestant Reformation?",
			Answer:          "Enabled mass distribution of Luther's theses and Bible translations",
			Category:        "history",
			HopCount:        2,
			RequiredSources: 3,
			Keywords:        []string{"printing press", "Protestant Reformation", "Luther"},
		},
	}
}

type SearchRequest struct {
	Query string `json:"query"`
	Mode  string `json:"mode"`
}

type SearchResponse struct {
	Answer    string   `json:"answer"`
	Sources   []Source `json:"sources"`
	Reasoning string   `json:"reasoning"`
}

type Source struct {
	Title       string  `json:"title"`
	URL         string  `json:"url"`
	Snippet     string  `json:"snippet"`
	Credibility float64 `json:"credibility"`
}

func runFRAMESQuestion(apiURL string, q FRAMESQuestion, mode string) FRAMESResult {
	start := time.Now()

	reqBody := SearchRequest{
		Query: q.Question,
		Mode:  mode,
	}
	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return FRAMESResult{
			Question:       q.Question,
			ExpectedAnswer: q.Answer,
			ActualAnswer:   fmt.Sprintf("ERROR: %v", err),
			Category:       q.Category,
			ProcessingTime: time.Since(start),
			Success:        false,
			Mode:           mode,
			HopCount:       q.HopCount,
		}
	}

	resp, err := http.Post(apiURL+"/api/search", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return FRAMESResult{
			Question:       q.Question,
			ExpectedAnswer: q.Answer,
			ActualAnswer:   fmt.Sprintf("ERROR: %v", err),
			Category:       q.Category,
			ProcessingTime: time.Since(start),
			Success:        false,
			Mode:           mode,
			HopCount:       q.HopCount,
		}
	}
	defer resp.Body.Close()

	processingTime := time.Since(start)

	if resp.StatusCode != http.StatusOK {
		return FRAMESResult{
			Question:       q.Question,
			ExpectedAnswer: q.Answer,
			ActualAnswer:   fmt.Sprintf("ERROR: HTTP %d", resp.StatusCode),
			Category:       q.Category,
			ProcessingTime: processingTime,
			Success:        false,
			Mode:           mode,
			HopCount:       q.HopCount,
		}
	}

	var result SearchResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return FRAMESResult{
			Question:       q.Question,
			ExpectedAnswer: q.Answer,
			ActualAnswer:   fmt.Sprintf("ERROR: %v", err),
			Category:       q.Category,
			ProcessingTime: processingTime,
			Success:        false,
			Mode:           mode,
			HopCount:       q.HopCount,
		}
	}

	// Evaluate metrics
	factuality := evaluateFactuality(result.Answer, q.Keywords)
	reasoningDepth := evaluateReasoningDepth(result.Reasoning, q.HopCount)
	sourceDiversity := evaluateSourceDiversity(result.Sources)

	success := result.Answer != "" &&
		len(result.Sources) >= q.RequiredSources &&
		factuality > 0.5

	return FRAMESResult{
		Question:        q.Question,
		ExpectedAnswer:  q.Answer,
		ActualAnswer:    result.Answer,
		Category:        q.Category,
		ProcessingTime:  processingTime,
		FactualityScore: factuality,
		ReasoningDepth:  reasoningDepth,
		SourceDiversity: sourceDiversity,
		SourceCount:     len(result.Sources),
		HopCount:        q.HopCount,
		Success:         success,
		Mode:            mode,
	}
}

func evaluateFactuality(answer string, keywords []string) float64 {
	answerLower := strings.ToLower(answer)
	matches := 0
	for _, keyword := range keywords {
		if strings.Contains(answerLower, strings.ToLower(keyword)) {
			matches++
		}
	}

	if len(keywords) == 0 {
		return 0.5
	}

	score := float64(matches) / float64(len(keywords))

	// Bonus for longer, detailed answers
	if len(answer) > 200 {
		score += 0.1
	}
	if len(answer) > 500 {
		score += 0.1
	}

	if score > 1.0 {
		score = 1.0
	}

	return score
}

func evaluateReasoningDepth(reasoning string, expectedHops int) float64 {
	if reasoning == "" {
		return 0.0
	}

	// Count reasoning steps
	steps := strings.Count(reasoning, "\n")
	if steps == 0 {
		steps = 1
	}

	// Compare with expected hops
	score := float64(steps) / float64(expectedHops*3) // Each hop ~3 steps

	if score > 1.0 {
		score = 1.0
	}

	return score
}

func evaluateSourceDiversity(sources []Source) float64 {
	if len(sources) == 0 {
		return 0.0
	}

	// Extract unique domains
	domains := make(map[string]bool)
	for _, src := range sources {
		// Simple domain extraction
		parts := strings.Split(src.URL, "/")
		if len(parts) > 2 {
			domain := parts[2]
			domains[domain] = true
		}
	}

	// Diversity = unique domains / total sources
	diversity := float64(len(domains)) / float64(len(sources))

	return diversity
}

func calculateFRAMESStats(results []FRAMESResult, totalTime time.Duration) FRAMESStats {
	stats := FRAMESStats{
		TotalQuestions: len(results),
		TotalTime:      totalTime,
	}

	var totalProcessingTime time.Duration
	var totalFactuality, totalDepth, totalDiversity float64

	for _, r := range results {
		if r.Success {
			stats.SuccessCount++
		} else {
			stats.FailCount++
		}
		totalProcessingTime += r.ProcessingTime
		totalFactuality += r.FactualityScore
		totalDepth += r.ReasoningDepth
		totalDiversity += r.SourceDiversity
	}

	if stats.TotalQuestions > 0 {
		stats.SuccessRate = float64(stats.SuccessCount) / float64(stats.TotalQuestions) * 100
		stats.AvgTime = totalProcessingTime.Seconds() / float64(stats.TotalQuestions)
		stats.AvgFactuality = totalFactuality / float64(stats.TotalQuestions)
		stats.AvgReasoningDepth = totalDepth / float64(stats.TotalQuestions)
		stats.AvgSourceDiv = totalDiversity / float64(stats.TotalQuestions)
	}

	return stats
}

func printFRAMESSummary(stats FRAMESStats) {
	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("           FRAMES BENCHMARK RESULTS")
	fmt.Println(strings.Repeat("=", 60))

	fmt.Printf("\n📊 Overall Statistics:\n")
	fmt.Printf("  Total Questions: %d\n", stats.TotalQuestions)
	fmt.Printf("  ✅ Success: %d\n", stats.SuccessCount)
	fmt.Printf("  ❌ Failed: %d\n", stats.FailCount)
	fmt.Printf("  Success Rate: %.2f%%\n", stats.SuccessRate)

	fmt.Printf("\n🎯 Quality Metrics:\n")
	fmt.Printf("  Avg Factuality Score: %.2f/1.0\n", stats.AvgFactuality)
	fmt.Printf("  Avg Reasoning Depth: %.2f/1.0\n", stats.AvgReasoningDepth)
	fmt.Printf("  Avg Source Diversity: %.2f/1.0\n", stats.AvgSourceDiv)

	fmt.Printf("\n⏱️  Performance:\n")
	fmt.Printf("  Average Time: %.2fs per question\n", stats.AvgTime)
	fmt.Printf("  Total Time: %.2fs\n", stats.TotalTime.Seconds())

	fmt.Println("\n" + strings.Repeat("=", 60))
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}

func saveFRAMESResults(results []FRAMESResult, filename string) error {
	data, err := json.MarshalIndent(results, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filename, data, 0o644)
}
