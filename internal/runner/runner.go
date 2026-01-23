package runner

import (
	"context"
	"log"
	"strings"
	"time"

	"codeworkers/internal/executor"
	"codeworkers/internal/models"
	"codeworkers/internal/queue"
)

const (
	StatusPassed = "PASSED"
	StatusFailed = "FAILED"

	// Log messages.
	LogWorkerStarted    = "Worker %d started"
	LogWorkerStopped    = "Worker %d stopped"
	LogWorkerProcessing = "Worker %d processing submission %s"
	LogWorkerCompleted  = "Worker %d completed submission %s with status: %s"
	LogRunningTestCase  = "Running test case %d for submission %s"
)

type Runner struct {
	queue    *queue.Queue
	executor *executor.DockerExecutor
	workers  int
	stop     chan struct{}
}

func NewRunner(q *queue.Queue, exec *executor.DockerExecutor, workers int) *Runner {
	return &Runner{
		queue:    q,
		executor: exec,
		workers:  workers,
		stop:     make(chan struct{}),
	}
}

func (r *Runner) Start() {
	for i := 0; i < r.workers; i++ {
		go r.worker(i)
	}
}

func (r *Runner) Stop() {
	close(r.stop)
}

func (r *Runner) worker(id int) {
	log.Printf(LogWorkerStarted, id)
	for {
		select {
		case <-r.stop:
			log.Printf(LogWorkerStopped, id)
			return
		case sub := <-r.queue.GetChannel():
			log.Printf(LogWorkerProcessing, id, sub.ID)
			result := r.execute(sub)
			r.queue.StoreResult(result)
			log.Printf(LogWorkerCompleted, id, sub.ID, result.Status)
		}
	}
}

func (r *Runner) execute(sub *models.Submission) *models.Result {
	result := &models.Result{
		SubmissionID: sub.ID,
		TestResults:  make([]models.TestResult, 0),
	}

	for i, testCase := range sub.TestCases {
		log.Printf(LogRunningTestCase, i+1, sub.ID)
		testResult := r.runTestCase(sub, testCase)
		result.TestResults = append(result.TestResults, testResult)

		if !testResult.Passed {
			result.Status = StatusFailed
			return result
		}
	}

	result.Status = StatusPassed
	return result
}

func (r *Runner) ExecuteSync(sub *models.Submission) *models.Result {
	return r.execute(sub)
}

func (r *Runner) runTestCase(sub *models.Submission, testCase models.TestCase) models.TestResult {
	start := time.Now()

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(sub.TimeLimit)*time.Second)
	defer cancel()

	output, err := r.executor.Execute(ctx, sub, testCase.Input)
	elapsed := time.Since(start).Seconds()

	if err != nil {
		return models.TestResult{
			Passed: false,
			Output: output,
			Error:  err.Error(),
			Time:   elapsed,
		}
	}

	passed := strings.TrimSpace(output) == strings.TrimSpace(testCase.Expected)
	return models.TestResult{
		Passed: passed,
		Output: output,
		Error:  "",
		Time:   elapsed,
	}
}
