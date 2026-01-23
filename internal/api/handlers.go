package api

import (
	"encoding/json"
	"log"
	"net/http"

	"codeworkers/internal/models"
	"codeworkers/internal/queue"
	"codeworkers/internal/runner"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

const (
	DefaultTimeLimit   = 5
	DefaultMemoryLimit = 256
	StatusQueued       = "QUEUED"
	StatusProcessing   = "PROCESSING"
	StatusPassed       = "PASSED"
	StatusFailed       = "FAILED"

	// Log messages.
	LogFailedWriteResponse       = "Failed to write response: %v"
	LogFailedWriteHealthResponse = "Failed to write health response: %v"
	LogSubmissionQueued          = "Submission %s queued"
	LogSubmissionExecuting       = "Executing submission %s synchronously"
	LogSubmissionCompleted       = "Submission %s completed with status: %s"
)

type API struct {
	queue  *queue.Queue
	runner *runner.Runner
}

func NewAPI(q *queue.Queue, r *runner.Runner) *API {
	return &API{queue: q, runner: r}
}

func (api *API) SetupRoutes() chi.Router {
	r := chi.NewRouter()

	r.Route("/api", func(r chi.Router) {
		r.Post("/submit", api.handleSubmit)
		r.Post("/submit/sync", api.handleSubmitSync)
		r.Get("/status/{id}", api.handleStatus)
	})

	r.Get("/health", api.handleHealth)

	return r
}

type SubmitRequest struct {
	Language    string            `json:"language"`
	Code        string            `json:"code"`
	TestCases   []models.TestCase `json:"test_cases"`
	TimeLimit   int               `json:"time_limit"`
	MemoryLimit int               `json:"memory_limit"`
}

type SubmitResponse struct {
	SubmissionID string `json:"submission_id"`
	Status       string `json:"status"`
}

type StatusResponse struct {
	SubmissionID string              `json:"submission_id"`
	Status       string              `json:"status"`
	TestResults  []models.TestResult `json:"test_results,omitempty"`
	Error        string              `json:"error,omitempty"`
}

func (api *API) handleSubmit(w http.ResponseWriter, r *http.Request) {
	submission, err := api.parseSubmitRequest(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	go api.queue.Enqueue(submission)

	resp := SubmitResponse{
		SubmissionID: submission.ID,
		Status:       StatusQueued,
	}

	if err := api.writeJSON(w, resp); err != nil {
		log.Printf(LogFailedWriteResponse, err)
	}
	log.Printf(LogSubmissionQueued, submission.ID)
}

func (api *API) handleSubmitSync(w http.ResponseWriter, r *http.Request) {
	submission, err := api.parseSubmitRequest(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	log.Printf(LogSubmissionExecuting, submission.ID)
	result := api.runner.ExecuteSync(submission)

	resp := StatusResponse{
		SubmissionID: result.SubmissionID,
		Status:       result.Status,
		TestResults:  result.TestResults,
		Error:        result.Error,
	}

	if err := api.writeJSON(w, resp); err != nil {
		log.Printf(LogFailedWriteResponse, err)
	}
	log.Printf(LogSubmissionCompleted, submission.ID, result.Status)
}

func (api *API) handleStatus(w http.ResponseWriter, r *http.Request) {
	submissionID := chi.URLParam(r, "id")
	if submissionID == "" {
		http.Error(w, "Submission ID required", http.StatusBadRequest)
		return
	}

	result, exists := api.queue.GetResult(submissionID)
	if !exists {
		resp := StatusResponse{
			SubmissionID: submissionID,
			Status:       StatusProcessing,
		}
		if err := api.writeJSON(w, resp); err != nil {
			log.Printf(LogFailedWriteResponse, err)
		}
		return
	}

	resp := StatusResponse{
		SubmissionID: result.SubmissionID,
		Status:       result.Status,
		TestResults:  result.TestResults,
		Error:        result.Error,
	}

	if err := api.writeJSON(w, resp); err != nil {
		log.Printf(LogFailedWriteResponse, err)
	}
}

func (api *API) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	if _, err := w.Write([]byte("OK")); err != nil {
		log.Printf(LogFailedWriteHealthResponse, err)
	}
}

func (api *API) parseSubmitRequest(r *http.Request) (*models.Submission, error) {
	var req SubmitRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return nil, err
	}

	if req.TimeLimit == 0 {
		req.TimeLimit = DefaultTimeLimit
	}
	if req.MemoryLimit == 0 {
		req.MemoryLimit = DefaultMemoryLimit
	}

	return &models.Submission{
		ID:          uuid.New().String(),
		Language:    req.Language,
		Code:        req.Code,
		TestCases:   req.TestCases,
		TimeLimit:   req.TimeLimit,
		MemoryLimit: req.MemoryLimit,
	}, nil
}

func (api *API) writeJSON(w http.ResponseWriter, data interface{}) error {
	w.Header().Set("Content-Type", "application/json")
	return json.NewEncoder(w).Encode(data)
}

func (api *API) Start(addr string) error {
	log.Printf("Starting API server on %s", addr)
	return http.ListenAndServe(addr, api.SetupRoutes())
}
