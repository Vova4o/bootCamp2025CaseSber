// benchmark/compare/main.go
package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
	"time"
)

type ComparisonResult struct {
	SimpleMode BenchmarkResult
	ProMode    BenchmarkResult
}

type BenchmarkResult struct {
	SimpleQA SimpleQAMetrics
	FRAMES   FRAMESMetrics
}

type SimpleQAMetrics struct {
	Accuracy float64
	AvgTime  float64
}

type FRAMESMetrics struct {
	SuccessRate     float64
	Factuality      float64
	ReasoningDepth  float64
	SourceDiversity float64
	AvgTime         float64
}

func main() {
	log.Println("🔬 Running Comprehensive Benchmark...")

	// Run SimpleQA for both modes
	log.Println("\n1️⃣ Testing Simple Mode on SimpleQA...")
	simpleQASimple := runSimpleQA("simple")

	log.Println("\n2️⃣ Testing Pro Mode on SimpleQA...")
	simpleQAPro := runSimpleQA("pro")

	// Run FRAMES for both modes
	log.Println("\n3️⃣ Testing Simple Mode on FRAMES...")
	framesSimple := runFRAMES("simple")

	log.Println("\n4️⃣ Testing Pro Mode on FRAMES...")
	framesPro := runFRAMES("pro")

	// Generate comparison report
	printComparison(simpleQASimple, simpleQAPro, framesSimple, framesPro)

	// Generate recommendation
	printRecommendation(simpleQASimple, simpleQAPro, framesSimple, framesPro)
}

func runSimpleQA(mode string) SimpleQAMetrics {
	outputFile := fmt.Sprintf("benchmark_%s_results.json", mode)

	cmd := exec.Command("go", "run", "./cmd/benchmark/simpleqa/main.go",
		"-mode", mode, "-limit", "10", "-output", outputFile)
	cmd.Dir = "/Users/vladimirgavrilenko/Pyproject/bootCamp2025CaseSber/backend-go"

	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Printf("Warning: SimpleQA %s mode failed: %v", mode, err)
		log.Printf("Output: %s", string(output))
		return SimpleQAMetrics{Accuracy: 0.0, AvgTime: 0.0}
	}

	log.Printf("SimpleQA %s mode output:\n%s", mode, string(output))

	// Parse JSON results
	data, err := os.ReadFile(outputFile)
	if err != nil {
		log.Printf("Warning: Failed to read results file: %v", err)
		return SimpleQAMetrics{Accuracy: 0.0, AvgTime: 0.0}
	}

	var results []SimpleQAResult
	if err := json.Unmarshal(data, &results); err != nil {
		log.Printf("Warning: Failed to parse results: %v", err)
		return SimpleQAMetrics{Accuracy: 0.0, AvgTime: 0.0}
	}

	// Calculate metrics
	correct := 0
	totalTime := 0.0
	for _, r := range results {
		if r.Correct {
			correct++
		}
		totalTime += r.ProcessingTime.Seconds()
	}

	accuracy := 0.0
	avgTime := 0.0
	if len(results) > 0 {
		accuracy = float64(correct) / float64(len(results)) * 100
		avgTime = totalTime / float64(len(results))
	}

	// Cleanup
	os.Remove(outputFile)

	return SimpleQAMetrics{
		Accuracy: accuracy,
		AvgTime:  avgTime,
	}
}

type SimpleQAResult struct {
	Question       string        `json:"question"`
	ExpectedAnswer string        `json:"expected_answer"`
	ActualAnswer   string        `json:"actual_answer"`
	Category       string        `json:"category"`
	ProcessingTime time.Duration `json:"processing_time"`
	Correct        bool          `json:"correct"`
	HasSources     bool          `json:"has_sources"`
	SourceCount    int           `json:"source_count"`
	Mode           string        `json:"mode"`
}

func runFRAMES(mode string) FRAMESMetrics {
	outputFile := fmt.Sprintf("frames_%s_results.json", mode)

	cmd := exec.Command("go", "run", "./cmd/benchmark/frames/main.go",
		"-mode", mode, "-limit", "5", "-output", outputFile)
	cmd.Dir = "/Users/vladimirgavrilenko/Pyproject/bootCamp2025CaseSber/backend-go"

	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Printf("Warning: FRAMES %s mode failed: %v", mode, err)
		log.Printf("Output: %s", string(output))
		return FRAMESMetrics{}
	}

	log.Printf("FRAMES %s mode output:\n%s", mode, string(output))

	// Parse JSON results
	data, err := os.ReadFile(outputFile)
	if err != nil {
		log.Printf("Warning: Failed to read FRAMES results: %v", err)
		return FRAMESMetrics{}
	}

	var results []FRAMESResult
	if err := json.Unmarshal(data, &results); err != nil {
		log.Printf("Warning: Failed to parse FRAMES results: %v", err)
		return FRAMESMetrics{}
	}

	// Calculate metrics
	successful := 0
	totalFactuality := 0.0
	totalReasoning := 0.0
	totalDiversity := 0.0
	totalTime := 0.0

	for _, r := range results {
		if r.Success {
			successful++
		}
		totalFactuality += r.FactualityScore
		totalReasoning += r.ReasoningDepth
		totalDiversity += r.SourceDiversity
		totalTime += r.ProcessingTime.Seconds()
	}

	metrics := FRAMESMetrics{}
	if len(results) > 0 {
		metrics.SuccessRate = float64(successful) / float64(len(results)) * 100
		metrics.Factuality = totalFactuality / float64(len(results))
		metrics.ReasoningDepth = totalReasoning / float64(len(results))
		metrics.SourceDiversity = totalDiversity / float64(len(results))
		metrics.AvgTime = totalTime / float64(len(results))
	}

	// Cleanup
	os.Remove(outputFile)

	return metrics
}

type FRAMESResult struct {
	Question        string        `json:"question"`
	ExpectedAnswer  string        `json:"expected_answer"`
	ActualAnswer    string        `json:"actual_answer"`
	HopCount        int           `json:"hop_count"`
	Success         bool          `json:"success"`
	FactualityScore float64       `json:"factuality_score"`
	ReasoningDepth  float64       `json:"reasoning_depth"`
	SourceDiversity float64       `json:"source_diversity"`
	ProcessingTime  time.Duration `json:"processing_time"`
	Mode            string        `json:"mode"`
}

func printComparison(simpleQASimple, simpleQAPro SimpleQAMetrics,
	framesSimple, framesPro FRAMESMetrics,
) {
	fmt.Println("\n" + strings.Repeat("=", 80))
	fmt.Println("                    📊 BENCHMARK COMPARISON")
	fmt.Println(strings.Repeat("=", 80))

	fmt.Println("\n🎯 SimpleQA (Factual Accuracy):")
	fmt.Printf("  %-20s  Simple: %.1f%%  |  Pro: %.1f%%  |  Δ: %+.1f%%\n",
		"Accuracy:", simpleQASimple.Accuracy, simpleQAPro.Accuracy,
		simpleQAPro.Accuracy-simpleQASimple.Accuracy)
	fmt.Printf("  %-20s  Simple: %.2fs  |  Pro: %.2fs  |  Δ: %+.2fs\n",
		"Avg Time:", simpleQASimple.AvgTime, simpleQAPro.AvgTime,
		simpleQAPro.AvgTime-simpleQASimple.AvgTime)

	fmt.Println("\n🔬 FRAMES (Multi-hop Reasoning):")
	fmt.Printf("  %-20s  Simple: %.1f%%  |  Pro: %.1f%%  |  Δ: %+.1f%%\n",
		"Success Rate:", framesSimple.SuccessRate, framesPro.SuccessRate,
		framesPro.SuccessRate-framesSimple.SuccessRate)
	fmt.Printf("  %-20s  Simple: %.2f   |  Pro: %.2f   |  Δ: %+.2f\n",
		"Factuality:", framesSimple.Factuality, framesPro.Factuality,
		framesPro.Factuality-framesSimple.Factuality)
	fmt.Printf("  %-20s  Simple: %.2f   |  Pro: %.2f   |  Δ: %+.2f\n",
		"Reasoning Depth:", framesSimple.ReasoningDepth, framesPro.ReasoningDepth,
		framesPro.ReasoningDepth-framesSimple.ReasoningDepth)
	fmt.Printf("  %-20s  Simple: %.2f   |  Pro: %.2f   |  Δ: %+.2f\n",
		"Source Diversity:", framesSimple.SourceDiversity, framesPro.SourceDiversity,
		framesPro.SourceDiversity-framesSimple.SourceDiversity)

	fmt.Println("\n" + strings.Repeat("=", 80))
}

func printRecommendation(simpleQASimple, simpleQAPro SimpleQAMetrics,
	framesSimple, framesPro FRAMESMetrics,
) {
	fmt.Println("\n💡 RECOMMENDATIONS:")
	fmt.Println(strings.Repeat("-", 80))

	fmt.Println("\n✅ Use Simple Mode when:")
	fmt.Println("  • Quick factual lookups (< 2s response time needed)")
	fmt.Println("  • Single-hop questions (Who? What? When?)")
	fmt.Println("  • Cost is a priority")
	fmt.Println("  • Accuracy > 90% is sufficient")

	fmt.Println("\n🚀 Use Pro Mode when:")
	fmt.Println("  • Complex multi-step reasoning required")
	fmt.Println("  • Need source verification and credibility scoring")
	fmt.Println("  • Comparison questions (Compare A vs B)")
	fmt.Println("  • Research and fact-checking scenarios")
	fmt.Println("  • Willing to trade speed for quality")

	fmt.Println("\n" + strings.Repeat("=", 80))
}
