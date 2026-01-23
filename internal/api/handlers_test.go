package api

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"codeworkers/internal/executor"
	"codeworkers/internal/models"
	"codeworkers/internal/queue"
	"codeworkers/internal/runner"

	"github.com/go-chi/chi/v5"
)

func setupTestAPI() *API {
	q := queue.NewQueue(10)
	registry := executor.NewStrategyRegistry()
	pool := executor.NewMockContainerPool()
	exec := executor.NewDockerExecutor(registry, pool)
	r := runner.NewRunner(q, exec, 2)
	return NewAPI(q, r)
}

func TestNewAPI(t *testing.T) {
	api := setupTestAPI()
	if api == nil {
		t.Fatal("NewAPI returned nil")
	}
	if api.queue == nil {
		t.Error("Queue not set")
	}
	if api.runner == nil {
		t.Error("Runner not set")
	}
}

func TestConstants(t *testing.T) {
	tests := []struct {
		name     string
		constant interface{}
		expected interface{}
	}{
		{"DefaultTimeLimit", DefaultTimeLimit, 5},
		{"DefaultMemoryLimit", DefaultMemoryLimit, 256},
		{"StatusQueued", StatusQueued, "QUEUED"},
		{"StatusProcessing", StatusProcessing, "PROCESSING"},
		{"StatusPassed", StatusPassed, "PASSED"},
		{"StatusFailed", StatusFailed, "FAILED"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.constant != tt.expected {
				t.Errorf("Expected %v, got %v", tt.expected, tt.constant)
			}
		})
	}
}

func TestLogConstants(t *testing.T) {
	constants := []string{
		LogFailedWriteResponse,
		LogFailedWriteHealthResponse,
		LogSubmissionQueued,
		LogSubmissionExecuting,
		LogSubmissionCompleted,
	}

	for _, c := range constants {
		if c == "" {
			t.Error("Log constant should not be empty")
		}
	}
}

func TestHandleHealth(t *testing.T) {
	api := setupTestAPI()
	req := httptest.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()

	api.handleHealth(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
	if w.Body.String() != "OK" {
		t.Errorf("Expected body 'OK', got %s", w.Body.String())
	}
}

func TestHandleSubmitInvalidJSON(t *testing.T) {
	api := setupTestAPI()
	req := httptest.NewRequest("POST", "/api/submit", bytes.NewBufferString("invalid json"))
	w := httptest.NewRecorder()

	api.handleSubmit(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}
}

func TestHandleSubmitValid(t *testing.T) {
	api := setupTestAPI()
	submitReq := SubmitRequest{
		Language: "python",
		Code:     "print('hello')",
		TestCases: []models.TestCase{
			{Input: "", Expected: "hello"},
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

	api.handleSubmit(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var resp SubmitResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if resp.SubmissionID == "" {
		t.Error("Expected non-empty submission ID")
	}
	if resp.Status != StatusQueued {
		t.Errorf("Expected status %s, got %s", StatusQueued, resp.Status)
	}
}

func TestHandleSubmitWithDefaults(t *testing.T) {
	api := setupTestAPI()
	submitReq := SubmitRequest{
		Language: "python",
		Code:     "print('hello')",
		TestCases: []models.TestCase{
			{Input: "", Expected: "hello"},
		},
	}

	body, err := json.Marshal(submitReq)
	if err != nil {
		t.Fatalf("Failed to marshal request: %v", err)
	}
	req := httptest.NewRequest("POST", "/api/submit", bytes.NewBuffer(body))
	w := httptest.NewRecorder()

	api.handleSubmit(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
}

func TestHandleStatusNotFound(t *testing.T) {
	api := setupTestAPI()
	req := httptest.NewRequest("GET", "/api/status/non-existent", nil)
	w := httptest.NewRecorder()

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "non-existent")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	api.handleStatus(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var resp StatusResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if resp.Status != StatusProcessing {
		t.Errorf("Expected status %s, got %s", StatusProcessing, resp.Status)
	}
}

func TestHandleStatusEmptyID(t *testing.T) {
	api := setupTestAPI()
	req := httptest.NewRequest("GET", "/api/status/", nil)
	w := httptest.NewRecorder()

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	api.handleStatus(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}
}

func TestSetupRoutes(t *testing.T) {
	api := setupTestAPI()
	router := api.SetupRoutes()

	if router == nil {
		t.Fatal("SetupRoutes returned nil")
	}
}

func TestParseSubmitRequest(t *testing.T) {
	api := setupTestAPI()
	submitReq := SubmitRequest{
		Language: "python",
		Code:     "print('hello')",
		TestCases: []models.TestCase{
			{Input: "", Expected: "hello"},
		},
	}

	body, err := json.Marshal(submitReq)
	if err != nil {
		t.Fatalf("Failed to marshal request: %v", err)
	}
	req := httptest.NewRequest("POST", "/api/submit", bytes.NewBuffer(body))

	submission, err := api.parseSubmitRequest(req)
	if err != nil {
		t.Fatalf("parseSubmitRequest failed: %v", err)
	}

	if submission.Language != submitReq.Language {
		t.Errorf("Expected language %s, got %s", submitReq.Language, submission.Language)
	}
	if submission.TimeLimit != DefaultTimeLimit {
		t.Errorf("Expected default time limit %d, got %d", DefaultTimeLimit, submission.TimeLimit)
	}
	if submission.MemoryLimit != DefaultMemoryLimit {
		t.Errorf("Expected default memory limit %d, got %d", DefaultMemoryLimit, submission.MemoryLimit)
	}
}

func TestWriteJSON(t *testing.T) {
	api := setupTestAPI()
	w := httptest.NewRecorder()

	data := map[string]string{"test": "value"}
	err := api.writeJSON(w, data)

	if err != nil {
		t.Errorf("writeJSON failed: %v", err)
	}

	contentType := w.Header().Get("Content-Type")
	if contentType != "application/json" {
		t.Errorf("Expected Content-Type application/json, got %s", contentType)
	}

	var result map[string]string
	if err := json.NewDecoder(w.Body).Decode(&result); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if result["test"] != "value" {
		t.Errorf("Expected value, got %s", result["test"])
	}
}

func TestHandleSubmitSyncValid(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping Docker test in short mode")
	}

	api := setupTestAPI()
	submitReq := SubmitRequest{
		Language: "python",
		Code:     "print('sync test')",
		TestCases: []models.TestCase{
			{Input: "", Expected: "sync test"},
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

	api.handleSubmitSync(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
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
}

func TestHandleSubmitSyncInvalidJSON(t *testing.T) {
	api := setupTestAPI()
	req := httptest.NewRequest("POST", "/api/submit/sync", bytes.NewBufferString("invalid"))
	w := httptest.NewRecorder()

	api.handleSubmitSync(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}
}

func TestHandleStatusWithResult(t *testing.T) {
	api := setupTestAPI()

	result := &models.Result{
		SubmissionID: "test-result-123",
		Status:       "PASSED",
		TestResults: []models.TestResult{
			{Passed: true, Output: "success", Time: 0.5},
		},
	}
	api.queue.StoreResult(result)

	req := httptest.NewRequest("GET", "/api/status/test-result-123", nil)
	w := httptest.NewRecorder()

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "test-result-123")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	api.handleStatus(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var resp StatusResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if resp.Status != "PASSED" {
		t.Errorf("Expected status PASSED, got %s", resp.Status)
	}
	if len(resp.TestResults) != 1 {
		t.Errorf("Expected 1 test result, got %d", len(resp.TestResults))
	}
}

func TestParseSubmitRequestInvalidJSON(t *testing.T) {
	api := setupTestAPI()
	req := httptest.NewRequest("POST", "/api/submit", bytes.NewBufferString("{invalid}"))

	_, err := api.parseSubmitRequest(req)
	if err == nil {
		t.Error("Expected error for invalid JSON")
	}
}

func TestParseSubmitRequestWithCustomLimits(t *testing.T) {
	api := setupTestAPI()
	submitReq := SubmitRequest{
		Language:    "python",
		Code:        "print('test')",
		TestCases:   []models.TestCase{{Input: "", Expected: "test"}},
		TimeLimit:   10,
		MemoryLimit: 512,
	}

	body, err := json.Marshal(submitReq)
	if err != nil {
		t.Fatalf("Failed to marshal request: %v", err)
	}
	req := httptest.NewRequest("POST", "/api/submit", bytes.NewBuffer(body))

	submission, err := api.parseSubmitRequest(req)
	if err != nil {
		t.Fatalf("parseSubmitRequest failed: %v", err)
	}

	if submission.TimeLimit != 10 {
		t.Errorf("Expected time limit 10, got %d", submission.TimeLimit)
	}
	if submission.MemoryLimit != 512 {
		t.Errorf("Expected memory limit 512, got %d", submission.MemoryLimit)
	}
}
