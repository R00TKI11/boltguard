package report

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/R00TKI11/boltguard/internal/facts"
	"github.com/R00TKI11/boltguard/internal/policy"
	"github.com/R00TKI11/boltguard/internal/rules"
)

// Report aggregates evaluation results for output
type Report struct {
	ImageName string
	Facts     *facts.Facts
	Results   []*rules.Result
	Policy    *policy.Policy
	Timestamp time.Time

	// computed stats
	TotalRules int
	Passed     int
	Failed     int
	BySeverity map[string]int
}

// New creates a report from evaluation results
func New(imageName string, f *facts.Facts, results []*rules.Result, p *policy.Policy) *Report {
	r := &Report{
		ImageName: imageName,
		Facts:     f,
		Results:   results,
		Policy:    p,
		Timestamp: time.Now(),
		TotalRules: len(results),
		BySeverity: rules.CountBySeverity(results),
	}

	for _, res := range results {
		if res.Passed {
			r.Passed++
		} else {
			r.Failed++
		}
	}

	return r
}

// Text outputs a human-readable text report
//nolint:errcheck // writes to stdout, nothing useful to do on error
func (r *Report) Text(w io.Writer) error {
	fmt.Fprintf(w, "BoltGuard Report\n")
	fmt.Fprintf(w, "================\n\n")

	fmt.Fprintf(w, "Image:    %s\n", r.ImageName)
	fmt.Fprintf(w, "Policy:   %s (v%s)\n", r.Policy.Name, r.Policy.Version)
	fmt.Fprintf(w, "Scanned:  %s\n\n", r.Timestamp.Format(time.RFC3339))

	// Summary
	fmt.Fprintf(w, "Summary\n")
	fmt.Fprintf(w, "-------\n")
	fmt.Fprintf(w, "Total checks: %d\n", r.TotalRules)
	fmt.Fprintf(w, "  Passed:     %d\n", r.Passed)
	fmt.Fprintf(w, "  Failed:     %d\n\n", r.Failed)

	if r.Failed > 0 {
		fmt.Fprintf(w, "Failures by severity:\n")
		if r.BySeverity["critical"] > 0 {
			fmt.Fprintf(w, "  Critical: %d\n", r.BySeverity["critical"])
		}
		if r.BySeverity["high"] > 0 {
			fmt.Fprintf(w, "  High:     %d\n", r.BySeverity["high"])
		}
		if r.BySeverity["medium"] > 0 {
			fmt.Fprintf(w, "  Medium:   %d\n", r.BySeverity["medium"])
		}
		if r.BySeverity["low"] > 0 {
			fmt.Fprintf(w, "  Low:      %d\n", r.BySeverity["low"])
		}
		fmt.Fprintf(w, "\n")
	}

	// Failures
	if r.Failed > 0 {
		fmt.Fprintf(w, "Failures\n")
		fmt.Fprintf(w, "--------\n")
		for _, res := range r.Results {
			if !res.Passed {
				fmt.Fprintf(w, "[%s] %s\n", strings.ToUpper(res.Severity), res.RuleName)
				fmt.Fprintf(w, "  ID:      %s\n", res.RuleID)
				fmt.Fprintf(w, "  Message: %s\n", res.Message)
				if res.Description != "" {
					fmt.Fprintf(w, "  Detail:  %s\n", res.Description)
				}
				fmt.Fprintf(w, "\n")
			}
		}
	}

	// Passed checks (brief)
	if r.Passed > 0 {
		fmt.Fprintf(w, "Passed Checks\n")
		fmt.Fprintf(w, "-------------\n")
		for _, res := range r.Results {
			if res.Passed {
				fmt.Fprintf(w, "✓ %s: %s\n", res.RuleName, res.Message)
			}
		}
		fmt.Fprintf(w, "\n")
	}

	// Final verdict
	if r.Failed == 0 {
		fmt.Fprintf(w, "✓ All checks passed\n")
		return nil
	}

	fmt.Fprintf(w, "✗ %d check(s) failed\n", r.Failed)
	return nil
}

// JSON outputs machine-readable JSON
func (r *Report) JSON(w io.Writer) error {
	output := map[string]interface{}{
		"image":     r.ImageName,
		"policy":    r.Policy.Name,
		"version":   r.Policy.Version,
		"timestamp": r.Timestamp.Format(time.RFC3339),
		"summary": map[string]interface{}{
			"total":       r.TotalRules,
			"passed":      r.Passed,
			"failed":      r.Failed,
			"by_severity": r.BySeverity,
		},
		"results": r.Results,
	}

	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(output)
}

// SARIF outputs SARIF format (static analysis results interchange format)
// https://docs.oasis-open.org/sarif/sarif/v2.1.0/sarif-v2.1.0.html
func (r *Report) SARIF(w io.Writer) error {
	sarif := map[string]interface{}{
		"version": "2.1.0",
		"$schema": "https://json.schemastore.org/sarif-2.1.0.json",
		"runs": []map[string]interface{}{
			{
				"tool": map[string]interface{}{
					"driver": map[string]interface{}{
						"name":            "BoltGuard",
						"informationUri":  "https://github.com/R00TKI11/boltguard",
						"version":         "0.1.0",
						"semanticVersion": "0.1.0",
						"rules":           r.buildSarifRules(),
					},
				},
				"artifacts": []map[string]interface{}{
					{
						"location": map[string]interface{}{
							"uri": r.ImageName,
						},
						"description": map[string]interface{}{
							"text": fmt.Sprintf("Container image: %s", r.ImageName),
						},
					},
				},
				"results": r.buildSarifResults(),
			},
		},
	}

	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(sarif)
}

func (r *Report) buildSarifRules() []map[string]interface{} {
	var sarifRules []map[string]interface{}

	for _, rule := range r.Policy.Rules {
		sarifRules = append(sarifRules, map[string]interface{}{
			"id":   rule.ID,
			"name": rule.Name,
			"shortDescription": map[string]string{
				"text": rule.Name,
			},
			"fullDescription": map[string]string{
				"text": rule.Description,
			},
			"defaultConfiguration": map[string]interface{}{
				"level": severityToLevel(rule.Severity),
			},
		})
	}

	return sarifRules
}

func (r *Report) buildSarifResults() []map[string]interface{} {
	var sarifResults []map[string]interface{}

	for _, res := range r.Results {
		if res.Passed {
			continue // SARIF typically only reports issues
		}

		sarifResults = append(sarifResults, map[string]interface{}{
			"ruleId": res.RuleID,
			"level":  severityToLevel(res.Severity),
			"message": map[string]string{
				"text": res.Message,
			},
			"locations": []map[string]interface{}{
				{
					"physicalLocation": map[string]interface{}{
						"artifactLocation": map[string]interface{}{
							"uri": r.ImageName,
						},
					},
				},
			},
		})
	}

	return sarifResults
}

func severityToLevel(severity string) string {
	switch severity {
	case "critical", "high":
		return "error"
	case "medium":
		return "warning"
	case "low":
		return "note"
	default:
		return "none"
	}
}
