package queue

import (
	"sync"

	"codeworkers/internal/models"
)

type Queue struct {
	submissions chan *models.Submission
	results     map[string]*models.Result
	mu          sync.RWMutex
}

func NewQueue(capacity int) *Queue {
	return &Queue{
		submissions: make(chan *models.Submission, capacity),
		results:     make(map[string]*models.Result),
	}
}

func (q *Queue) Enqueue(sub *models.Submission) {
	q.submissions <- sub
}

func (q *Queue) Dequeue() *models.Submission {
	return <-q.submissions
}

func (q *Queue) GetChannel() <-chan *models.Submission {
	return q.submissions
}

func (q *Queue) StoreResult(result *models.Result) {
	q.mu.Lock()
	defer q.mu.Unlock()
	q.results[result.SubmissionID] = result
}

func (q *Queue) GetResult(id string) (*models.Result, bool) {
	q.mu.RLock()
	defer q.mu.RUnlock()
	result, exists := q.results[id]
	return result, exists
}
