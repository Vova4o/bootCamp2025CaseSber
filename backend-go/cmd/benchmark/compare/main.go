// benchmark/compare/main.go
package main

import (
	"fmt"
	"log"
	"os/exec"
	"strings"
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
	cmd := exec.Command("go", "run", "../simpleqa/main.go", 
		"-mode", mode, "-limit", "10")
	output, _ := cmd.CombinedOutput()
	
	// Parse output...
	return SimpleQAMetrics{
		Accuracy: 0.0, // Parse from output
		AvgTime:  0.0,
	}
}

func runFRAMES(mode string) FRAMESMetrics {
	cmd := exec.Command("go", "run", "../frames/main.go", 
		"-mode", mode, "-limit", "5")
	output, _ := cmd.CombinedOutput()
	
	// Parse output...
	return FRAMESMetrics{}
}

func printComparison(simpleQASimple, simpleQAPro SimpleQAMetrics, 
	framesSimple, framesPro FRAMESMetrics) {
	
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
	framesSimple, framesPro FRAMESMetrics) {
	
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