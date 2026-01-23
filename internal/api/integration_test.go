package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"codeworkers/internal/executor"
	"codeworkers/internal/models"
	"codeworkers/internal/queue"
	"codeworkers/internal/runner"
)

func TestIntegrationSubmitAndStatus(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	q := queue.NewQueue(10)
	registry := executor.NewStrategyRegistry()
	pool := executor.NewMockContainerPool()
	exec := executor.NewDockerExecutor(registry, pool)
	r := runner.NewRunner(q, exec, 2)
	r.Start()
	defer r.Stop()

	api := NewAPI(q, r)
	router := api.SetupRoutes()

	submitReq := SubmitRequest{
		Language: "python",
		Code:     "print(input())",
		TestCases: []models.TestCase{
			{Input: "hello", Expected: "hello"},
		},
		TimeLimit:   5,
		MemoryLimit: 256,
	}

	body, err := json.Marshal(submitReq)
	if err != nil {
		t.Fatalf("Failed to marshal request: %v", err)
	}
	req := httptest.NewRequest("POST", "/api/submit", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Submit failed with status %d", w.Code)
	}

	var submitResp SubmitResponse
	if err := json.NewDecoder(w.Body).Decode(&submitResp); err != nil {
		t.Fatalf("Failed to decode submit response: %v", err)
	}

	submissionID := submitResp.SubmissionID
	if submissionID == "" {
		t.Fatal("No submission ID returned")
	}

	time.Sleep(2 * time.Second)

	statusReq := httptest.NewRequest("GET", "/api/status/"+submissionID, nil)
	statusW := httptest.NewRecorder()

	router.ServeHTTP(statusW, statusReq)

	if statusW.Code != http.StatusOK {
		t.Fatalf("Status check failed with status %d", statusW.Code)
	}

	var statusResp StatusResponse
	if err := json.NewDecoder(statusW.Body).Decode(&statusResp); err != nil {
		t.Fatalf("Failed to decode status response: %v", err)
	}

	if statusResp.SubmissionID != submissionID {
		t.Errorf("Expected submission ID %s, got %s", submissionID, statusResp.SubmissionID)
	}
}

func TestIntegrationSyncSubmit(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	q := queue.NewQueue(10)
	registry := executor.NewStrategyRegistry()
	pool := executor.NewMockContainerPool()
	exec := executor.NewDockerExecutor(registry, pool)
	r := runner.NewRunner(q, exec, 2)

	api := NewAPI(q, r)
	router := api.SetupRoutes()

	submitReq := SubmitRequest{
		Language: "python",
		Code:     "n = int(input())\nprint(n * 2)",
		TestCases: []models.TestCase{
			{Input: "5", Expected: "10"},
			{Input: "10", Expected: "20"},
		},
		TimeLimit:   5,
		MemoryLimit: 256,
	}

	body, err := json.Marshal(submitReq)
	if err != nil {
		t.Fatalf("Failed to marshal request: %v", err)
	}
	req := httptest.NewRequest("POST", "/api/submit/sync", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Sync submit failed with status %d: %s", w.Code, w.Body.String())
	}

	var resp StatusResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if resp.SubmissionID == "" {
		t.Error("Expected non-empty submission ID")
	}
	if resp.Status == "" {
		t.Error("Expected non-empty status")
	}
	if len(resp.TestResults) != 2 {
		t.Errorf("Expected 2 test results, got %d", len(resp.TestResults))
	}
}

func TestIntegrationHealthCheck(t *testing.T) {
	api := setupTestAPI()
	router := api.SetupRoutes()

	req := httptest.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
	if w.Body.String() != "OK" {
		t.Errorf("Expected body 'OK', got %s", w.Body.String())
	}
}

func TestIntegrationMultipleLanguages(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	q := queue.NewQueue(10)
	registry := executor.NewStrategyRegistry()
	pool := executor.NewMockContainerPool()
	exec := executor.NewDockerExecutor(registry, pool)
	r := runner.NewRunner(q, exec, 2)

	api := NewAPI(q, r)
	router := api.SetupRoutes()

	tests := []struct {
		language string
		code     string
		input    string
		expected string
	}{
		{
			language: "python",
			code:     "print(input())",
			input:    "test",
			expected: "test",
		},
		{
			language: "javascript",
			code:     "const input = require('fs').readFileSync(0, 'utf-8').trim(); console.log(input);",
			input:    "test",
			expected: "test",
		},
	}

	for _, tt := range tests {
		t.Run(tt.language, func(t *testing.T) {
			submitReq := SubmitRequest{
				Language: tt.language,
				Code:     tt.code,
				TestCases: []models.TestCase{
					{Input: tt.input, Expected: tt.expected},
				},
				TimeLimit:   5,
				MemoryLimit: 256,
			}

			body, err := json.Marshal(submitReq)
			if err != nil {
				t.Fatalf("Failed to marshal request: %v", err)
			}
			req := httptest.NewRequest("POST", "/api/submit/sync", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			if w.Code != http.StatusOK {
				t.Fatalf("Submit failed with status %d: %s", w.Code, w.Body.String())
			}
		})
	}
}
