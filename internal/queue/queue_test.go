package queue

import (
	"sync"
	"testing"

	"codeworkers/internal/models"
)

func TestNewQueue(t *testing.T) {
	capacity := 10
	q := NewQueue(capacity)

	if q == nil {
		t.Fatal("NewQueue returned nil")
	}
	if q.submissions == nil {
		t.Error("submissions channel not initialized")
	}
	if q.results == nil {
		t.Error("results map not initialized")
	}
}

func TestQueueEnqueueDequeue(t *testing.T) {
	q := NewQueue(5)
	sub := &models.Submission{
		ID:       "test-1",
		Language: "python",
		Code:     "print('hello')",
	}

	go q.Enqueue(sub)
	result := q.Dequeue()

	if result.ID != sub.ID {
		t.Errorf("Expected ID %s, got %s", sub.ID, result.ID)
	}
}

func TestQueueStoreAndGetResult(t *testing.T) {
	q := NewQueue(5)
	result := &models.Result{
		SubmissionID: "test-1",
		Status:       "PASSED",
	}

	q.StoreResult(result)

	retrieved, exists := q.GetResult("test-1")
	if !exists {
		t.Error("Result not found")
	}
	if retrieved.SubmissionID != result.SubmissionID {
		t.Errorf("Expected ID %s, got %s", result.SubmissionID, retrieved.SubmissionID)
	}
	if retrieved.Status != result.Status {
		t.Errorf("Expected status %s, got %s", result.Status, retrieved.Status)
	}
}

func TestQueueGetResultNotFound(t *testing.T) {
	q := NewQueue(5)

	_, exists := q.GetResult("non-existent")
	if exists {
		t.Error("Expected result not to exist")
	}
}

func TestQueueConcurrency(t *testing.T) {
	q := NewQueue(100)
	var wg sync.WaitGroup

	// Concurrent writes
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			result := &models.Result{
				SubmissionID: string(rune(id)),
				Status:       "PASSED",
			}
			q.StoreResult(result)
		}(i)
	}

	// Concurrent reads
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			q.GetResult(string(rune(id)))
		}(i)
	}

	wg.Wait()
}

func TestGetChannel(t *testing.T) {
	q := NewQueue(5)
	ch := q.GetChannel()

	if ch == nil {
		t.Error("GetChannel returned nil")
	}

	sub := &models.Submission{ID: "test"}
	go q.Enqueue(sub)

	received := <-ch
	if received.ID != sub.ID {
		t.Errorf("Expected ID %s, got %s", sub.ID, received.ID)
	}
}
