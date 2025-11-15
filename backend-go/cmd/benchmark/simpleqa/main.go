package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

// ============================================================================
// API Types
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
// SimpleQA Dataset Types (Hugging Face format)
// ============================================================================

type SimpleQAMetadata struct {
	Topic      string   `json:"topic"`
	AnswerType string   `json:"answer_type"`
	URLs       []string `json:"urls"`
}

// Raw row from HF API - metadata is a JSON string
type SimpleQARowRaw struct {
	MetadataStr string `json:"metadata"` // Comes as JSON string
	Problem     string `json:"problem"`
	Answer      string `json:"answer"`
}

type HuggingFaceResponse struct {
	Rows []struct {
		Row SimpleQARowRaw `json:"row"`
	} `json:"rows"`
	NumRowsTotal int `json:"num_rows_total"`
}

// Unified question format
type BenchmarkQuestion struct {
	ID         string
	Question   string
	Answer     string
	Category   string
	AnswerType string
	URLs       []string
	Dataset    string
}

// ============================================================================
// Result Types
// ============================================================================

type BenchmarkResult struct {
	ID                string        `json:"id"`
	Question          string        `json:"question"`
	ExpectedAnswer    string        `json:"expected_answer"`
	ActualAnswer      string        `json:"actual_answer"`
	Category          string        `json:"category"`
	AnswerType        string        `json:"answer_type"`
	Dataset           string        `json:"dataset"`
	Mode              string        `json:"mode"`
	ProcessingTime    time.Duration `json:"processing_time"`
	Correct           bool          `json:"correct"`
	PartiallyCorrect  bool          `json:"partially_correct"`
	HasSources        bool          `json:"has_sources"`
	SourceCount       int           `json:"source_count"`
	SourceQuality     float64       `json:"source_quality"`
	FactualityScore   float64       `json:"factuality_score"`
	Error             string        `json:"error,omitempty"`
}

type CategoryStats struct {
	Total            int
	Correct          int
	PartiallyCorrect int
	Accuracy         float64
	PartialAccuracy  float64
	AvgTime          float64
	AvgSources       float64
}

type Stats struct {
	TotalQuestions     int
	CorrectCount       int
	PartialCount       int
	FailCount          int
	Accuracy           float64
	PartialAccuracy    float64
	AvgTime            float64
	AvgSourceCount     float64
	AvgFactualityScore float64
	TotalTime          time.Duration
	ByCategory         map[string]CategoryStats
	ByAnswerType       map[string]CategoryStats
}

// ============================================================================
// Main
// ============================================================================

func main() {
	mode := flag.String("mode", "simple", "Mode: simple or pro")
	limit := flag.Int("limit", 100, "Number of questions (0 = all)")
	offset := flag.Int("offset", 0, "Starting offset in dataset")
	output := flag.String("output", "", "Output file (auto-generated if empty)")
	apiURL := flag.String("api", "http://localhost:8000", "Backend API URL")
	hfToken := flag.String("hf-token", "", "Hugging Face API token (optional)")
	useLocal := flag.Bool("local", false, "Use local dataset file")
	localFile := flag.String("file", "simpleqa_dataset.json", "Local dataset file")
	flag.Parse()

	log.Printf("🧪 SimpleQA Benchmark - Research Assistant")
	log.Printf("   Mode: %s | API: %s", *mode, *apiURL)

	// Load questions
	var questions []BenchmarkQuestion
	var err error

	if *useLocal {
		questions, err = loadLocalDataset(*localFile)
	} else {
		questions, err = loadSimpleQAFromHF(*hfToken, *offset, *limit)
	}

	if err != nil {
		log.Fatalf("❌ Failed to load dataset: %v", err)
	}

	log.Printf("✅ Loaded %d questions from SimpleQA dataset", len(questions))

	// Run benchmark
	startTime := time.Now()
	results := runBenchmark(*apiURL, questions, *mode)
	totalTime := time.Since(startTime)

	// Calculate statistics
	stats := calculateStats(results, totalTime)

	// Print summary
	printDetailedSummary(stats, *mode)

	// Save results
	if *output == "" {
		*output = fmt.Sprintf("simpleqa_benchmark_%s_%s.json",
			*mode, time.Now().Format("20060102_150405"))
	}
	if err := saveResults(results, stats, *output); err != nil {
		log.Printf("⚠️  Warning: Failed to save results: %v", err)
	} else {
		log.Printf("💾 Results saved to %s", *output)
	}
}

// ============================================================================
// Dataset Loading from Hugging Face
// ============================================================================

func loadSimpleQAFromHF(token string, offset, limit int) ([]BenchmarkQuestion, error) {
	if limit == 0 {
		limit = 4326 // Total rows in dataset
	}

	// Hugging Face datasets API endpoint
	url := fmt.Sprintf(
		"https://datasets-server.huggingface.co/rows?dataset=basicv8vc/SimpleQA&config=default&split=test&offset=%d&length=%d",
		offset, limit,
	)

	log.Printf("📡 Fetching from Hugging Face: offset=%d, limit=%d", offset, limit)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("HF API error %d: %s", resp.StatusCode, body)
	}

	var hfResponse HuggingFaceResponse
	if err := json.NewDecoder(resp.Body).Decode(&hfResponse); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	log.Printf("📊 Total rows in dataset: %d", hfResponse.NumRowsTotal)

	questions := make([]BenchmarkQuestion, 0, len(hfResponse.Rows))
	for i, row := range hfResponse.Rows {
		// Parse metadata JSON string
		metadata, err := parseMetadata(row.Row.MetadataStr)
		if err != nil {
			log.Printf("⚠️  Warning: Failed to parse metadata for row %d: %v", i, err)
			continue
		}

		questions = append(questions, BenchmarkQuestion{
			ID:         fmt.Sprintf("simpleqa_%d", offset+i+1),
			Question:   row.Row.Problem,
			Answer:     row.Row.Answer,
			Category:   metadata.Topic,
			AnswerType: metadata.AnswerType,
			URLs:       metadata.URLs,
			Dataset:    "simpleqa",
		})
	}

	return questions, nil
}

// Parse metadata from JSON string (or Python dict string)
func parseMetadata(metadataStr string) (SimpleQAMetadata, error) {
	var metadata SimpleQAMetadata

	// Replace Python-style single quotes with double quotes
	metadataStr = strings.ReplaceAll(metadataStr, "'", "\"")

	if err := json.Unmarshal([]byte(metadataStr), &metadata); err != nil {
		return metadata, fmt.Errorf("failed to parse metadata: %w", err)
	}

	return metadata, nil
}

func loadLocalDataset(filename string) ([]BenchmarkQuestion, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		if os.IsNotExist(err) {
			log.Printf("⚠️  Dataset file not found, creating sample...")
			return createSampleDataset(), nil
		}
		return nil, err
	}

	var rows []struct {
		Metadata interface{} `json:"metadata"`
		Problem  string      `json:"problem"`
		Answer   string      `json:"answer"`
	}

	if err := json.Unmarshal(data, &rows); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}

	questions := make([]BenchmarkQuestion, 0, len(rows))
	for i, row := range rows {
		var metadata SimpleQAMetadata

		// Handle metadata as either string or object
		switch v := row.Metadata.(type) {
		case string:
			if parsed, err := parseMetadata(v); err == nil {
				metadata = parsed
			}
		case map[string]interface{}:
			// Convert to JSON and back
			if jsonBytes, err := json.Marshal(v); err == nil {
				json.Unmarshal(jsonBytes, &metadata)
			}
		}

		questions = append(questions, BenchmarkQuestion{
			ID:         fmt.Sprintf("simpleqa_%d", i+1),
			Question:   row.Problem,
			Answer:     row.Answer,
			Category:   metadata.Topic,
			AnswerType: metadata.AnswerType,
			URLs:       metadata.URLs,
			Dataset:    "simpleqa",
		})
	}

	return questions, nil
}

func createSampleDataset() []BenchmarkQuestion {
	return []BenchmarkQuestion{
		{
			ID:         "sample_1",
			Question:   "What is the capital of France?",
			Answer:     "Paris",
			Category:   "Geography",
			AnswerType: "Place",
			Dataset:    "simpleqa",
		},
		{
			ID:         "sample_2",
			Question:   "Who wrote Romeo and Juliet?",
			Answer:     "William Shakespeare",
			Category:   "Art",
			AnswerType: "Person",
			Dataset:    "simpleqa",
		},
	}
}

// ============================================================================
// Benchmark Execution
// ============================================================================

func runBenchmark(apiURL string, questions []BenchmarkQuestion, mode string) []BenchmarkResult {
	results := make([]BenchmarkResult, 0, len(questions))

	for i, q := range questions {
		log.Printf("\n[%d/%d] ❓ %s", i+1, len(questions), truncate(q.Question, 100))
		log.Printf("  📌 Expected: %s", truncate(q.Answer, 80))
		log.Printf("  🏷️  Category: %s | Type: %s", q.Category, q.AnswerType)

		result := runQuestion(apiURL, q, mode)
		results = append(results, result)

		status := "✅"
		if result.PartiallyCorrect {
			status = "🟡"
		} else if !result.Correct {
			status = "❌"
		}

		log.Printf("  💬 Got: %s", truncate(result.ActualAnswer, 80))
		log.Printf("  %s %s | ⏱️  %.2fs | 📚 %d sources | ✓ %.2f",
			status,
			formatResult(result),
			result.ProcessingTime.Seconds(),
			result.SourceCount,
			result.FactualityScore)
	}

	return results
}

func runQuestion(apiURL string, q BenchmarkQuestion, mode string) BenchmarkResult {
	start := time.Now()

	reqBody := SearchRequest{
		Query: q.Question,
		Mode:  mode,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return createErrorResult(q, mode, err, time.Since(start))
	}

	resp, err := http.Post(apiURL+"/api/search", "application/json",
		bytes.NewBuffer(jsonData))
	if err != nil {
		return createErrorResult(q, mode, err, time.Since(start))
	}
	defer resp.Body.Close()

	processingTime := time.Since(start)

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return createErrorResult(q, mode,
			fmt.Errorf("HTTP %d: %s", resp.StatusCode, body), processingTime)
	}

	var searchResp SearchResponse
	if err := json.NewDecoder(resp.Body).Decode(&searchResp); err != nil {
		return createErrorResult(q, mode, err, processingTime)
	}

	// Evaluate result
	correct, partial := evaluateAnswer(searchResp.Answer, q.Answer)
	sourceQuality := evaluateSourceQuality(searchResp.Sources, q.URLs)
	factualityScore := evaluateFactuality(searchResp.Answer, q.Answer)

	return BenchmarkResult{
		ID:               q.ID,
		Question:         q.Question,
		ExpectedAnswer:   q.Answer,
		ActualAnswer:     searchResp.Answer,
		Category:         q.Category,
		AnswerType:       q.AnswerType,
		Dataset:          q.Dataset,
		Mode:             mode,
		ProcessingTime:   processingTime,
		Correct:          correct,
		PartiallyCorrect: partial,
		HasSources:       len(searchResp.Sources) > 0,
		SourceCount:      len(searchResp.Sources),
		SourceQuality:    sourceQuality,
		FactualityScore:  factualityScore,
	}
}

func createErrorResult(q BenchmarkQuestion, mode string, err error, duration time.Duration) BenchmarkResult {
	return BenchmarkResult{
		ID:             q.ID,
		Question:       q.Question,
		ExpectedAnswer: q.Answer,
		ActualAnswer:   "",
		Category:       q.Category,
		AnswerType:     q.AnswerType,
		Dataset:        q.Dataset,
		Mode:           mode,
		ProcessingTime: duration,
		Correct:        false,
		Error:          err.Error(),
	}
}

// ============================================================================
// Evaluation Functions
// ============================================================================

func evaluateAnswer(actual, expected string) (correct, partial bool) {
	if actual == "" {
		return false, false
	}

	actualLower := strings.ToLower(strings.TrimSpace(actual))
	expectedLower := strings.ToLower(strings.TrimSpace(expected))

	// Exact substring match
	if strings.Contains(actualLower, expectedLower) {
		return true, false
	}

	// Check if expected is contained in actual
	if strings.Contains(expectedLower, actualLower) {
		return true, false
	}

	// Extract key terms (words longer than 3 chars)
	expectedWords := extractKeyWords(expectedLower)
	actualWords := extractKeyWords(actualLower)

	// Count matches
	matchCount := 0
	for _, expWord := range expectedWords {
		for _, actWord := range actualWords {
			if expWord == actWord {
				matchCount++
				break
			}
		}
	}

	// Partial match if >50% of key words match
	if len(expectedWords) > 0 {
		matchRatio := float64(matchCount) / float64(len(expectedWords))
		if matchRatio >= 0.8 {
			return true, false
		}
		if matchRatio >= 0.5 {
			return false, true
		}
	}

	return false, false
}

func extractKeyWords(text string) []string {
	words := strings.Fields(text)
	keyWords := make([]string, 0)

	stopWords := map[string]bool{
		"the": true, "is": true, "at": true, "which": true, "on": true,
		"and": true, "or": true, "but": true, "in": true, "with": true,
		"was": true, "were": true, "been": true, "being": true, "a": true,
		"an": true, "of": true, "to": true, "for": true, "as": true,
	}

	for _, word := range words {
		cleaned := strings.Trim(word, ".,!?;:\"'()[]{}«»")
		if len(cleaned) > 3 && !stopWords[cleaned] {
			keyWords = append(keyWords, cleaned)
		}
	}

	return keyWords
}

func evaluateSourceQuality(sources []Source, expectedURLs []string) float64 {
	if len(sources) == 0 {
		return 0.0
	}

	score := 0.0

	// Base score for having sources
	score += 0.4

	// Diversity score (unique domains)
	domains := make(map[string]bool)
	for _, s := range sources {
		domain := extractDomain(s.URL)
		domains[domain] = true
	}
	diversityScore := float64(len(domains)) / float64(len(sources))
	score += diversityScore * 0.3

	// Expected URL matching (if provided)
	if len(expectedURLs) > 0 {
		matches := 0
		for _, expectedURL := range expectedURLs {
			expectedDomain := extractDomain(expectedURL)
			for _, actual := range sources {
				actualDomain := extractDomain(actual.URL)
				if strings.Contains(actualDomain, expectedDomain) ||
					strings.Contains(expectedDomain, actualDomain) {
					matches++
					break
				}
			}
		}
		score += (float64(matches) / float64(len(expectedURLs))) * 0.3
	} else {
		score += 0.3
	}

	return score
}

func extractDomain(url string) string {
	// Remove protocol
	url = strings.TrimPrefix(url, "http://")
	url = strings.TrimPrefix(url, "https://")

	// Extract domain
	parts := strings.Split(url, "/")
	if len(parts) > 0 {
		// Remove www. if present
		domain := strings.TrimPrefix(parts[0], "www.")
		return domain
	}
	return url
}

func evaluateFactuality(actual, expected string) float64 {
	if actual == "" {
		return 0.0
	}

	actualWords := extractKeyWords(strings.ToLower(actual))
	expectedWords := extractKeyWords(strings.ToLower(expected))

	if len(expectedWords) == 0 {
		return 0.0
	}

	matches := 0
	for _, expWord := range expectedWords {
		for _, actWord := range actualWords {
			if expWord == actWord {
				matches++
				break
			}
		}
	}

	return float64(matches) / float64(len(expectedWords))
}

// ============================================================================
// Statistics
// ============================================================================

func calculateStats(results []BenchmarkResult, totalTime time.Duration) Stats {
	stats := Stats{
		TotalQuestions: len(results),
		TotalTime:      totalTime,
		ByCategory:     make(map[string]CategoryStats),
		ByAnswerType:   make(map[string]CategoryStats),
	}

	var totalProcessingTime time.Duration
	var totalSources, totalFactuality float64

	for _, r := range results {
		if r.Correct {
			stats.CorrectCount++
		} else if r.PartiallyCorrect {
			stats.PartialCount++
		} else {
			stats.FailCount++
		}

		totalProcessingTime += r.ProcessingTime
		totalSources += float64(r.SourceCount)
		totalFactuality += r.FactualityScore

		// By category
		updateCategoryStats(stats.ByCategory, r.Category, r)
		// By answer type
		updateCategoryStats(stats.ByAnswerType, r.AnswerType, r)
	}

	if stats.TotalQuestions > 0 {
		stats.Accuracy = float64(stats.CorrectCount) /
			float64(stats.TotalQuestions) * 100
		stats.PartialAccuracy = float64(stats.CorrectCount+stats.PartialCount) /
			float64(stats.TotalQuestions) * 100
		stats.AvgTime = totalProcessingTime.Seconds() /
			float64(stats.TotalQuestions)
		stats.AvgSourceCount = totalSources / float64(stats.TotalQuestions)
		stats.AvgFactualityScore = totalFactuality / float64(stats.TotalQuestions)
	}

	// Finalize category stats
	finalizeStatsMap(stats.ByCategory)
	finalizeStatsMap(stats.ByAnswerType)

	return stats
}

func updateCategoryStats(statsMap map[string]CategoryStats, key string, r BenchmarkResult) {
	cat := statsMap[key]
	cat.Total++
	if r.Correct {
		cat.Correct++
	}
	if r.PartiallyCorrect {
		cat.PartiallyCorrect++
	}
	cat.AvgTime += r.ProcessingTime.Seconds()
	cat.AvgSources += float64(r.SourceCount)
	statsMap[key] = cat
}

func finalizeStatsMap(statsMap map[string]CategoryStats) {
	for key, catStats := range statsMap {
		if catStats.Total > 0 {
			catStats.Accuracy = float64(catStats.Correct) /
				float64(catStats.Total) * 100
			catStats.PartialAccuracy = float64(catStats.Correct+catStats.PartiallyCorrect) /
				float64(catStats.Total) * 100
			catStats.AvgTime /= float64(catStats.Total)
			catStats.AvgSources /= float64(catStats.Total)
		}
		statsMap[key] = catStats
	}
}

// ============================================================================
// Output
// ============================================================================

func printDetailedSummary(stats Stats, mode string) {
	fmt.Println("\n" + strings.Repeat("=", 70))
	fmt.Printf("      SIMPLEQA BENCHMARK RESULTS\n")
	fmt.Printf("      Mode: %s\n", strings.ToUpper(mode))
	fmt.Println(strings.Repeat("=", 70))

	fmt.Printf("\n📊 Overall Performance:\n")
	fmt.Printf("  Total Questions: %d\n", stats.TotalQuestions)
	fmt.Printf("  ✅ Fully Correct: %d (%.1f%%)\n",
		stats.CorrectCount, stats.Accuracy)
	fmt.Printf("  🟡 Partially Correct: %d\n", stats.PartialCount)
	fmt.Printf("  ❌ Incorrect: %d\n", stats.FailCount)
	fmt.Printf("  🎯 Strict Accuracy: %.2f%%\n", stats.Accuracy)
	fmt.Printf("  🎯 Lenient Accuracy: %.2f%%\n", stats.PartialAccuracy)

	fmt.Printf("\n📚 Quality Metrics:\n")
	fmt.Printf("  📖 Avg Sources: %.1f per question\n", stats.AvgSourceCount)
	fmt.Printf("  ✓ Avg Factuality Score: %.2f\n", stats.AvgFactualityScore)

	fmt.Printf("\n⏱️  Performance:\n")
	fmt.Printf("  Average Time: %.2fs per question\n", stats.AvgTime)
	fmt.Printf("  Total Time: %.2fs\n", stats.TotalTime.Seconds())

	if len(stats.ByCategory) > 0 {
		fmt.Printf("\n📂 By Category:\n")
		for cat, catStats := range stats.ByCategory {
			icon := getAccuracyIcon(catStats.Accuracy)
			fmt.Printf("  %s %-25s: %.1f%% (%d/%d) | ⏱️  %.2fs | 📚 %.1f\n",
				icon, cat, catStats.Accuracy, catStats.Correct, catStats.Total,
				catStats.AvgTime, catStats.AvgSources)
		}
	}

	if len(stats.ByAnswerType) > 0 {
		fmt.Printf("\n🔤 By Answer Type:\n")
		for ansType, typeStats := range stats.ByAnswerType {
			icon := getAccuracyIcon(typeStats.Accuracy)
			fmt.Printf("  %s %-25s: %.1f%% (%d/%d)\n",
				icon, ansType, typeStats.Accuracy, typeStats.Correct, typeStats.Total)
		}
	}

	fmt.Println("\n" + strings.Repeat("=", 70))
}

func getAccuracyIcon(accuracy float64) string {
	if accuracy >= 80 {
		return "✅"
	} else if accuracy >= 50 {
		return "🟡"
	}
	return "❌"
}

func formatResult(r BenchmarkResult) string {
	if r.Correct {
		return "CORRECT"
	} else if r.PartiallyCorrect {
		return "PARTIAL"
	}
	return "INCORRECT"
}

func truncate(s string, length int) string {
	if len(s) <= length {
		return s
	}
	return s[:length] + "..."
}

func saveResults(results []BenchmarkResult, stats Stats, filename string) error {
	output := struct {
		Timestamp string            `json:"timestamp"`
		Stats     Stats             `json:"stats"`
		Results   []BenchmarkResult `json:"results"`
	}{
		Timestamp: time.Now().Format(time.RFC3339),
		Stats:     stats,
		Results:   results,
	}

	data, err := json.MarshalIndent(output, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filename, data, 0644)
}