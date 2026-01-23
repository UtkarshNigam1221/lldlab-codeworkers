package models

type Submission struct {
	ID          string
	Language    string
	Code        string
	TestCases   []TestCase
	TimeLimit   int
	MemoryLimit int
}

type TestCase struct {
	Input    string
	Expected string
}

type Result struct {
	SubmissionID string
	Status       string
	TestResults  []TestResult
	Error        string
}

type TestResult struct {
	Passed bool
	Output string
	Error  string
	Time   float64
}
