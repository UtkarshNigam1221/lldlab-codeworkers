package executor

import (
	"strings"
	"testing"
)

func TestPythonStrategy(t *testing.T) {
	strategy := &PythonStrategy{}

	if strategy.GetFilename() != "solution.py" {
		t.Errorf("Expected solution.py, got %s", strategy.GetFilename())
	}

	code := "print('hello')"
	input := "test"
	cmd := strategy.BuildRunCommand(code, input)

	if !strings.Contains(cmd, "python3") {
		t.Error("Command should contain python3")
	}
	if !strings.Contains(cmd, "solution.py") {
		t.Error("Command should contain solution.py")
	}
}

func TestCppStrategy(t *testing.T) {
	strategy := &CppStrategy{}

	if strategy.GetFilename() != "solution.cpp" {
		t.Errorf("Expected solution.cpp, got %s", strategy.GetFilename())
	}

	code := "int main() {}"
	input := ""
	cmd := strategy.BuildRunCommand(code, input)

	if !strings.Contains(cmd, "g++") {
		t.Error("Command should contain g++")
	}
	if !strings.Contains(cmd, "./solution") {
		t.Error("Command should contain ./solution")
	}
}

func TestJavaStrategy(t *testing.T) {
	strategy := &JavaStrategy{}

	if strategy.GetFilename() != "Solution.java" {
		t.Errorf("Expected Solution.java, got %s", strategy.GetFilename())
	}

	code := "public class Solution {}"
	input := ""
	cmd := strategy.BuildRunCommand(code, input)

	if !strings.Contains(cmd, "javac") {
		t.Error("Command should contain javac")
	}
	if !strings.Contains(cmd, "java Solution") {
		t.Error("Command should contain java Solution")
	}
}

func TestJavaScriptStrategy(t *testing.T) {
	strategy := &JavaScriptStrategy{}

	if strategy.GetFilename() != "solution.js" {
		t.Errorf("Expected solution.js, got %s", strategy.GetFilename())
	}

	code := "console.log('test')"
	input := ""
	cmd := strategy.BuildRunCommand(code, input)

	if !strings.Contains(cmd, "node") {
		t.Error("Command should contain node")
	}
}

func TestTypeScriptStrategy(t *testing.T) {
	strategy := &TypeScriptStrategy{}

	if strategy.GetFilename() != "solution.ts" {
		t.Errorf("Expected solution.ts, got %s", strategy.GetFilename())
	}

	code := "console.log('test')"
	input := ""
	cmd := strategy.BuildRunCommand(code, input)

	if !strings.Contains(cmd, "ts-node") {
		t.Error("Command should contain ts-node")
	}
}

func TestGoStrategy(t *testing.T) {
	strategy := &GoStrategy{}

	if strategy.GetFilename() != "solution.go" {
		t.Errorf("Expected solution.go, got %s", strategy.GetFilename())
	}

	code := "package main"
	input := ""
	cmd := strategy.BuildRunCommand(code, input)

	if !strings.Contains(cmd, "go run") {
		t.Error("Command should contain go run")
	}
}

func TestStrategyRegistry(t *testing.T) {
	registry := NewStrategyRegistry()

	tests := []struct {
		language string
		wantErr  bool
	}{
		{"python", false},
		{"javascript", false},
		{"typescript", false},
		{"cpp", false},
		{"java", false},
		{"go", false},
		{"rust", true},
		{"unknown", true},
	}

	for _, tt := range tests {
		t.Run(tt.language, func(t *testing.T) {
			strategy, err := registry.GetStrategy(tt.language)
			if tt.wantErr {
				if err == nil {
					t.Error("Expected error for unsupported language")
				}
				if strategy != nil {
					t.Error("Expected nil strategy for unsupported language")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if strategy == nil {
					t.Error("Expected strategy, got nil")
				}
			}
		})
	}
}

func TestEscapeString(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"hello", "hello"},
		{"hello'world", "hello'\\''world"},
		{"it's", "it'\\''s"},
		{"multiple'single'quotes", "multiple'\\''single'\\''quotes"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := escapeString(tt.input)
			if result != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, result)
			}
		})
	}
}
