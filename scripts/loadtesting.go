package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"sort"
	"sync"
	"time"
)

type TestPayload struct {
	Language    string     `json:"language"`
	Code        string     `json:"code"`
	TestCases   []TestCase `json:"test_cases"`
	TimeLimit   int        `json:"time_limit"`
	MemoryLimit int        `json:"memory_limit"`
}

type TestCase struct {
	Input    string `json:"input"`
	Expected string `json:"expected"`
}

type RequestResult struct {
	StatusCode int
	Duration   time.Duration
	Error      error
	Language   string
}

type TestMetrics struct {
	Concurrent     int
	Total          int
	Success        int
	Failed         int
	SuccessRate    float64
	RequestsPerSec float64
	AvgDuration    time.Duration
	MedianDuration time.Duration
	MinDuration    time.Duration
	MaxDuration    time.Duration
	LanguageStats  map[string]int
}

type LanguageTest struct {
	Language string
	Code     string
	Input    string
	Expected string
}

var languageTests = []LanguageTest{
	{Language: "python", Code: "print(input())", Input: "hello", Expected: "hello"},
	{Language: "javascript", Code: "console.log(require('fs').readFileSync(0, 'utf-8').trim())", Input: "world", Expected: "world"},
	{Language: "java", Code: "import java.util.Scanner;\npublic class Main {\n    public static void main(String[] args) {\n        Scanner sc = new Scanner(System.in);\n        System.out.println(sc.nextLine());\n    }\n}", Input: "test", Expected: "test"},
	{Language: "go", Code: "package main\nimport (\n\"bufio\"\n\"fmt\"\n\"os\"\n)\nfunc main() {\nscanner := bufio.NewScanner(os.Stdin)\nscanner.Scan()\nfmt.Println(scanner.Text())\n}", Input: "golang", Expected: "golang"},
	{Language: "cpp", Code: "#include <iostream>\n#include <string>\nusing namespace std;\nint main() {\n    string s;\n    getline(cin, s);\n    cout << s << endl;\n    return 0;\n}", Input: "cpp", Expected: "cpp"},
}

func main() {
	host := flag.String("host", "http://54.179.130.165:8080", "API host")
	maxConcurrent := flag.Int("max", 50, "Maximum concurrent requests")
	step := flag.Int("step", 5, "Concurrency step increment")
	requestsPerTest := flag.Int("requests", 20, "Number of requests per test (ignored if duration is set)")
	duration := flag.Int("duration", 0, "Test duration in seconds (0 = use request count)")
	randomLang := flag.Bool("random-lang", false, "Use random languages for each request")
	flag.Parse()

	rand.Seed(time.Now().UnixNano())

	fmt.Println("🔥 Load Testing Codeworkers API")
	fmt.Println("=" + repeat("=", 49))
	fmt.Printf("Host: %s\n", *host)
	fmt.Printf("Max Concurrent: %d\n", *maxConcurrent)
	fmt.Printf("Step: %d\n", *step)
	if *duration > 0 {
		fmt.Printf("Test duration: %ds (sustained load)\n", *duration)
	} else {
		fmt.Printf("Requests per test: %d\n", *requestsPerTest)
	}
	fmt.Printf("Random languages: %v\n", *randomLang)
	fmt.Println()

	// Health check
	fmt.Println("🔍 Checking if service is up...")
	if !checkHealth(*host) {
		fmt.Println("❌ Service is not reachable")
		return
	}
	fmt.Println("✅ Service is up")
	fmt.Println()

	// Run tests
	fmt.Println("🚀 Starting load tests...")
	fmt.Println()

	var allResults []TestMetrics

	for concurrent := 20; concurrent <= *maxConcurrent; concurrent += *step {
		fmt.Printf("📊 Testing with %d concurrent requests...\n", concurrent)

		var metrics TestMetrics
		if *duration > 0 {
			metrics = runSustainedLoadTest(*host, concurrent, *duration, *randomLang)
		} else {
			metrics = runLoadTest(*host, concurrent, *requestsPerTest, *randomLang)
		}
		allResults = append(allResults, metrics)

		// Display results
		fmt.Printf("   ✅ Success: %d/%d\n", metrics.Success, metrics.Total)
		fmt.Printf("   ❌ Failed: %d\n", metrics.Failed)
		fmt.Printf("   📈 Success Rate: %.1f%%\n", metrics.SuccessRate)
		fmt.Printf("   ⚡ Requests/sec: %.2f\n", metrics.RequestsPerSec)
		fmt.Printf("   ⏱️  Avg Response: %dms\n", metrics.AvgDuration.Milliseconds())
		fmt.Printf("   ⏱️  Median Response: %dms\n", metrics.MedianDuration.Milliseconds())
		fmt.Printf("   ⏱️  Min/Max: %dms / %dms\n", metrics.MinDuration.Milliseconds(), metrics.MaxDuration.Milliseconds())
		if *randomLang && len(metrics.LanguageStats) > 0 {
			fmt.Printf("   🌐 Languages tested: ")
			for lang, count := range metrics.LanguageStats {
				fmt.Printf("%s(%d) ", lang, count)
			}
			fmt.Println()
		}
		fmt.Println()

		// Cool down between tests
		time.Sleep(2 * time.Second)
	}

	// Summary
	fmt.Println()
	fmt.Println("📋 Load Test Summary")
	fmt.Println("=" + repeat("=", 49))
	fmt.Printf("%-12s %-10s %-10s %-10s %-10s %-10s\n", "Concurrent", "Success", "Failed", "Rate%", "RPS", "Avg(ms)")
	fmt.Println(repeat("-", 50))

	for _, r := range allResults {
		fmt.Printf("%-12d %-10d %-10d %-10.1f %-10.2f %-10d\n",
			r.Concurrent, r.Success, r.Failed, r.SuccessRate, r.RequestsPerSec, r.AvgDuration.Milliseconds())
	}

	fmt.Println()
	fmt.Println("💡 Analysis:")

	// Find optimal concurrency (95%+ success rate)
	var optimal *TestMetrics
	for i := range allResults {
		if allResults[i].SuccessRate >= 95 {
			if optimal == nil || allResults[i].Concurrent > optimal.Concurrent {
				optimal = &allResults[i]
			}
		}
	}

	if optimal != nil {
		fmt.Printf("   ✅ Recommended max concurrent: %d (%.1f%% success, %.1f req/s)\n",
			optimal.Concurrent, optimal.SuccessRate, optimal.RequestsPerSec)
		fmt.Printf("   📊 Max sustained TPS: %.1f\n", optimal.RequestsPerSec)
	} else {
		fmt.Println("   ⚠️  Service struggled at all tested concurrency levels")
	}

	// Find breaking point
	for _, r := range allResults {
		if r.SuccessRate < 80 {
			fmt.Printf("   ⚠️  Performance degrades at %d concurrent requests\n", r.Concurrent)
			break
		}
	}

	fmt.Println()
	fmt.Println("✅ Load test complete!")
}

func checkHealth(host string) bool {
	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Get(host + "/health")
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	return resp.StatusCode == 200
}

func getRandomLanguageTest() LanguageTest {
	return languageTests[rand.Intn(len(languageTests))]
}

func runLoadTest(host string, concurrent, totalRequests int, randomLang bool) TestMetrics {
	url := host + "/api/submit/sync"

	startTime := time.Now()
	results := make(chan RequestResult, totalRequests)

	// Semaphore to limit concurrency
	sem := make(chan struct{}, concurrent)
	var wg sync.WaitGroup

	// Launch requests
	for i := 0; i < totalRequests; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			sem <- struct{}{}        // Acquire
			defer func() { <-sem }() // Release

			var langTest LanguageTest
			if randomLang {
				langTest = getRandomLanguageTest()
			} else {
				langTest = languageTests[0] // Default to Python
			}

			payload := TestPayload{
				Language: langTest.Language,
				Code:     langTest.Code,
				TestCases: []TestCase{
					{Input: langTest.Input, Expected: langTest.Expected},
				},
				TimeLimit:   5,
				MemoryLimit: 256,
			}

			payloadBytes, _ := json.Marshal(payload)
			result := makeRequest(url, payloadBytes, langTest.Language)
			results <- result
		}()

		// Small delay to spread requests
		time.Sleep(10 * time.Millisecond)
	}

	// Wait for all requests to complete
	wg.Wait()
	close(results)

	elapsed := time.Since(startTime)

	// Process results
	var durations []time.Duration
	success := 0
	failed := 0
	langStats := make(map[string]int)

	for result := range results {
		if result.StatusCode == 200 {
			success++
			durations = append(durations, result.Duration)
			langStats[result.Language]++
		} else {
			failed++
		}
	}

	// Calculate metrics
	metrics := TestMetrics{
		Concurrent:     concurrent,
		Total:          totalRequests,
		Success:        success,
		Failed:         failed,
		SuccessRate:    float64(success) / float64(totalRequests) * 100,
		RequestsPerSec: float64(totalRequests) / elapsed.Seconds(),
		LanguageStats:  langStats,
	}

	if len(durations) > 0 {
		sort.Slice(durations, func(i, j int) bool {
			return durations[i] < durations[j]
		})

		var sum time.Duration
		for _, d := range durations {
			sum += d
		}

		metrics.AvgDuration = sum / time.Duration(len(durations))
		metrics.MedianDuration = durations[len(durations)/2]
		metrics.MinDuration = durations[0]
		metrics.MaxDuration = durations[len(durations)-1]
	}

	return metrics
}

func runSustainedLoadTest(host string, concurrent int, durationSec int, randomLang bool) TestMetrics {
	url := host + "/api/submit/sync"

	startTime := time.Now()
	deadline := startTime.Add(time.Duration(durationSec) * time.Second)
	results := make(chan RequestResult, 1000)

	// Semaphore to limit concurrency
	sem := make(chan struct{}, concurrent)
	var wg sync.WaitGroup

	// Launch continuous requests until deadline
	for time.Now().Before(deadline) {
		wg.Add(1)
		go func() {
			defer wg.Done()
			sem <- struct{}{}        // Acquire
			defer func() { <-sem }() // Release

			var langTest LanguageTest
			if randomLang {
				langTest = getRandomLanguageTest()
			} else {
				langTest = languageTests[0]
			}

			payload := TestPayload{
				Language: langTest.Language,
				Code:     langTest.Code,
				TestCases: []TestCase{
					{Input: langTest.Input, Expected: langTest.Expected},
				},
				TimeLimit:   5,
				MemoryLimit: 256,
			}

			payloadBytes, _ := json.Marshal(payload)
			result := makeRequest(url, payloadBytes, langTest.Language)
			results <- result
		}()

		// Small delay to maintain steady rate
		time.Sleep(10 * time.Millisecond)
	}

	// Wait for all requests to complete
	wg.Wait()
	close(results)

	elapsed := time.Since(startTime)

	// Process results
	var durations []time.Duration
	success := 0
	failed := 0
	langStats := make(map[string]int)
	totalRequests := 0

	for result := range results {
		totalRequests++
		if result.StatusCode == 200 {
			success++
			durations = append(durations, result.Duration)
			langStats[result.Language]++
		} else {
			failed++
		}
	}

	// Calculate metrics
	metrics := TestMetrics{
		Concurrent:     concurrent,
		Total:          totalRequests,
		Success:        success,
		Failed:         failed,
		SuccessRate:    float64(success) / float64(totalRequests) * 100,
		RequestsPerSec: float64(totalRequests) / elapsed.Seconds(),
		LanguageStats:  langStats,
	}

	if len(durations) > 0 {
		sort.Slice(durations, func(i, j int) bool {
			return durations[i] < durations[j]
		})

		var sum time.Duration
		for _, d := range durations {
			sum += d
		}

		metrics.AvgDuration = sum / time.Duration(len(durations))
		metrics.MedianDuration = durations[len(durations)/2]
		metrics.MinDuration = durations[0]
		metrics.MaxDuration = durations[len(durations)-1]
	}

	return metrics
}

func makeRequest(url string, payload []byte, language string) RequestResult {
	client := &http.Client{Timeout: 30 * time.Second}

	start := time.Now()
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(payload))
	if err != nil {
		return RequestResult{StatusCode: 0, Duration: time.Since(start), Error: err, Language: language}
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	duration := time.Since(start)

	if err != nil {
		return RequestResult{StatusCode: 0, Duration: duration, Error: err, Language: language}
	}
	defer resp.Body.Close()

	// Read and discard body
	io.Copy(io.Discard, resp.Body)

	return RequestResult{StatusCode: resp.StatusCode, Duration: duration, Error: nil, Language: language}
}

func repeat(s string, count int) string {
	result := ""
	for i := 0; i < count; i++ {
		result += s
	}
	return result
}
