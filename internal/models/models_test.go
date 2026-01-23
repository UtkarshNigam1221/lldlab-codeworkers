package models

import "testing"

func TestSubmissionStructure(t *testing.T) {
	sub := &Submission{
		ID:          "test-123",
		Language:    "python",
		Code:        "print('hello')",
		TestCases:   []TestCase{{Input: "test", Expected: "output"}},
		TimeLimit:   5,
		MemoryLimit: 256,
	}

	if sub.ID != "test-123" {
		t.Errorf("Expected ID test-123, got %s", sub.ID)
	}
	if sub.Language != "python" {
		t.Errorf("Expected language python, got %s", sub.Language)
	}
	if len(sub.TestCases) != 1 {
		t.Errorf("Expected 1 test case, got %d", len(sub.TestCases))
	}
}

func TestTestCaseStructure(t *testing.T) {
	tc := TestCase{
		Input:    "5",
		Expected: "10",
	}

	if tc.Input != "5" {
		t.Errorf("Expected input 5, got %s", tc.Input)
	}
	if tc.Expected != "10" {
		t.Errorf("Expected output 10, got %s", tc.Expected)
	}
}

func TestResultStructure(t *testing.T) {
	result := &Result{
		SubmissionID: "test-456",
		Status:       "PASSED",
		TestResults:  []TestResult{{Passed: true, Output: "success", Time: 0.5}},
		Error:        "",
	}

	if result.SubmissionID != "test-456" {
		t.Errorf("Expected submission ID test-456, got %s", result.SubmissionID)
	}
	if result.Status != "PASSED" {
		t.Errorf("Expected status PASSED, got %s", result.Status)
	}
	if len(result.TestResults) != 1 {
		t.Errorf("Expected 1 test result, got %d", len(result.TestResults))
	}
}

func TestTestResultStructure(t *testing.T) {
	tr := TestResult{
		Passed: true,
		Output: "hello world",
		Error:  "",
		Time:   1.23,
	}

	if !tr.Passed {
		t.Error("Expected test to be passed")
	}
	if tr.Output != "hello world" {
		t.Errorf("Expected output 'hello world', got %s", tr.Output)
	}
	if tr.Time != 1.23 {
		t.Errorf("Expected time 1.23, got %f", tr.Time)
	}
}

func TestSubmissionWithMultipleTestCases(t *testing.T) {
	sub := &Submission{
		ID:       "test-multi",
		Language: "javascript",
		Code:     "console.log(input)",
		TestCases: []TestCase{
			{Input: "1", Expected: "1"},
			{Input: "2", Expected: "2"},
			{Input: "3", Expected: "3"},
		},
		TimeLimit:   10,
		MemoryLimit: 512,
	}

	if len(sub.TestCases) != 3 {
		t.Errorf("Expected 3 test cases, got %d", len(sub.TestCases))
	}

	for i, tc := range sub.TestCases {
		if tc.Input != tc.Expected {
			t.Errorf("Test case %d: input and expected should match", i)
		}
	}
}

func TestResultWithError(t *testing.T) {
	result := &Result{
		SubmissionID: "test-error",
		Status:       "FAILED",
		TestResults: []TestResult{
			{Passed: false, Output: "", Error: "Runtime error", Time: 0.1},
		},
		Error: "Compilation failed",
	}

	if result.Status != "FAILED" {
		t.Errorf("Expected status FAILED, got %s", result.Status)
	}
	if result.Error == "" {
		t.Error("Expected non-empty error message")
	}
	if result.TestResults[0].Passed {
		t.Error("Expected test to be failed")
	}
}

func TestEmptyStructures(t *testing.T) {
	sub := &Submission{}
	if sub.ID != "" {
		t.Error("New submission should have empty ID")
	}

	result := &Result{}
	if result.Status != "" {
		t.Error("New result should have empty status")
	}

	testCase := TestCase{}
	if testCase.Input != "" || testCase.Expected != "" {
		t.Error("New test case should have empty fields")
	}

	testResult := TestResult{}
	if testResult.Passed {
		t.Error("New test result should not be passed by default")
	}
}
