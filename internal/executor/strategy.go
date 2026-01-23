package executor

import (
	"fmt"
	"strings"
)

type LanguageStrategy interface {
	GetFilename() string
	BuildRunCommand(code, input string) string
}

type PythonStrategy struct{}

func (s *PythonStrategy) GetFilename() string {
	return "solution.py"
}

func (s *PythonStrategy) BuildRunCommand(code, input string) string {
	return fmt.Sprintf("echo '%s' > %s && echo '%s' | python3 %s",
		escapeString(code), s.GetFilename(), escapeString(input), s.GetFilename())
}

type CppStrategy struct{}

func (s *CppStrategy) GetFilename() string {
	return "solution.cpp"
}

func (s *CppStrategy) BuildRunCommand(code, input string) string {
	return fmt.Sprintf("echo '%s' > %s && g++ -o solution %s && echo '%s' | ./solution",
		escapeString(code), s.GetFilename(), s.GetFilename(), escapeString(input))
}

type JavaStrategy struct{}

func (s *JavaStrategy) GetFilename() string {
	return "Solution.java"
}

func (s *JavaStrategy) BuildRunCommand(code, input string) string {
	return fmt.Sprintf("echo '%s' > %s && javac %s && echo '%s' | java Solution",
		escapeString(code), s.GetFilename(), s.GetFilename(), escapeString(input))
}

type JavaScriptStrategy struct{}

func (s *JavaScriptStrategy) GetFilename() string {
	return "solution.js"
}

func (s *JavaScriptStrategy) BuildRunCommand(code, input string) string {
	return fmt.Sprintf("echo '%s' > %s && echo '%s' | node %s",
		escapeString(code), s.GetFilename(), escapeString(input), s.GetFilename())
}

type TypeScriptStrategy struct{}

func (s *TypeScriptStrategy) GetFilename() string {
	return "solution.ts"
}

func (s *TypeScriptStrategy) BuildRunCommand(code, input string) string {
	return fmt.Sprintf("echo '%s' > %s && npx ts-node %s <<< '%s'",
		escapeString(code), s.GetFilename(), s.GetFilename(), escapeString(input))
}

type GoStrategy struct{}

func (s *GoStrategy) GetFilename() string {
	return "solution.go"
}

func (s *GoStrategy) BuildRunCommand(code, input string) string {
	return fmt.Sprintf("echo '%s' > %s && go run %s <<< '%s'",
		escapeString(code), s.GetFilename(), s.GetFilename(), escapeString(input))
}

type StrategyRegistry struct {
	strategies map[string]LanguageStrategy
}

func NewStrategyRegistry() *StrategyRegistry {
	return &StrategyRegistry{
		strategies: map[string]LanguageStrategy{
			"python":     &PythonStrategy{},
			"javascript": &JavaScriptStrategy{},
			"typescript": &TypeScriptStrategy{},
			"cpp":        &CppStrategy{},
			"java":       &JavaStrategy{},
			"go":         &GoStrategy{},
		},
	}
}

func (r *StrategyRegistry) GetStrategy(language string) (LanguageStrategy, error) {
	strategy, exists := r.strategies[language]
	if !exists {
		return nil, fmt.Errorf("unsupported language: %s", language)
	}
	return strategy, nil
}

func escapeString(s string) string {
	s = strings.ReplaceAll(s, "'", "'\\''")
	return s
}
