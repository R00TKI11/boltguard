package rules

import (
	"fmt"

	"github.com/yourusername/boltguard/internal/facts"
	"github.com/yourusername/boltguard/internal/policy"
)

// Result represents the outcome of evaluating a single rule
type Result struct {
	RuleID      string
	RuleName    string
	Severity    string
	Passed      bool
	Message     string
	Description string
}

// Engine evaluates rules against facts
type Engine struct {
	evaluators map[string]Evaluator
}

// Evaluator is the interface all rule types must implement
type Evaluator interface {
	Evaluate(facts *facts.Facts, rule *policy.Rule) (*Result, error)
}

// NewEngine creates a rules engine with all built-in evaluators
func NewEngine() *Engine {
	e := &Engine{
		evaluators: make(map[string]Evaluator),
	}

	// register all evaluators
	e.Register("user", &UserEvaluator{})
	e.Register("size", &SizeEvaluator{})
	e.Register("label", &LabelEvaluator{})
	e.Register("env", &EnvEvaluator{})
	e.Register("base", &BaseEvaluator{})
	e.Register("layers", &LayersEvaluator{})

	return e
}

// Register adds a custom evaluator
func (e *Engine) Register(kind string, eval Evaluator) {
	e.evaluators[kind] = eval
}

// Evaluate runs all policy rules against the facts
func (e *Engine) Evaluate(f *facts.Facts, p *policy.Policy) []*Result {
	var results []*Result

	for _, rule := range p.Rules {
		evaluator, exists := e.evaluators[rule.Kind]
		if !exists {
			// unknown rule kind - skip or error depending on policy settings
			results = append(results, &Result{
				RuleID:   rule.ID,
				RuleName: rule.Name,
				Severity: rule.Severity,
				Passed:   false,
				Message:  fmt.Sprintf("unknown rule kind: %s", rule.Kind),
			})
			continue
		}

		result, err := evaluator.Evaluate(f, &rule)
		if err != nil {
			results = append(results, &Result{
				RuleID:   rule.ID,
				RuleName: rule.Name,
				Severity: rule.Severity,
				Passed:   false,
				Message:  fmt.Sprintf("evaluation error: %v", err),
			})
			continue
		}

		// fill in metadata from rule
		result.RuleID = rule.ID
		result.RuleName = rule.Name
		result.Severity = rule.Severity
		result.Description = rule.Description

		results = append(results, result)
	}

	return results
}

// CountFailures returns the number of failed checks
func CountFailures(results []*Result) int {
	count := 0
	for _, r := range results {
		if !r.Passed {
			count++
		}
	}
	return count
}

// CountBySeverity counts failures grouped by severity
func CountBySeverity(results []*Result) map[string]int {
	counts := map[string]int{
		"critical": 0,
		"high":     0,
		"medium":   0,
		"low":      0,
		"info":     0,
	}

	for _, r := range results {
		if !r.Passed {
			counts[r.Severity]++
		}
	}

	return counts
}

// HasCriticalFailures checks if any critical rules failed
func HasCriticalFailures(results []*Result) bool {
	for _, r := range results {
		if !r.Passed && r.Severity == "critical" {
			return true
		}
	}
	return false
}
